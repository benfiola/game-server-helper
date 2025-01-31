package helper

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"slices"
	"time"
)

// fileCacheVersion is used to ensure that the on-disk file cache manifest uses an up-to-date schema
const fileCacheVersion = "1"

// fileCacheFetchCb is called during a 'put' to populate a destination path
type fileCacheFetchCb func(string) error

// fileCacheItem is an item within the cache
type fileCacheItem struct {
	IsFile       bool      `json:"isFile"`
	Key          string    `json:"key"`
	LastAccessed time.Time `json:"lastAccessed"`
	LastUuid     string    `json:"lastUuid"`
	Path         string    `json:"path"`
	Size         int       `json:"size"`
}

// fileCache holds all the metadata associated with the file cache
type fileCache struct {
	contents  Map[string, fileCacheItem]
	ctx       context.Context
	dir       string
	logger    *slog.Logger
	sizeLimit int
	uuid      string
}

// fileCacheManifest represents file cache state persisted on-disk
type fileCacheManifest struct {
	Contents map[string]fileCacheItem `json:"contents"`
	Version  string                   `json:"version"`
}

// Cleans the [fileCache] by removing untracked files and non-existent files from the cache directory.
// Also ensures non-existent records are removed from the cache contents map.
// Returns an error if the clean operation fails
func (fc *fileCache) clean() error {
	defer fc.save()
	validPaths := map[string]bool{}
	for key, item := range fc.contents {
		_, err := os.Lstat(item.Path)
		if err == nil {
			validPaths[item.Path] = true
			continue
		}
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		fc.logger.Info("remove missing cache item", "key", key, "path", item.Path)
		err = RemovePaths(fc.ctx, item.Path)
		if err != nil {
			return err
		}
		delete(fc.contents, key)
	}
	subpaths, err := ListDir(fc.ctx, fc.dir)
	if err != nil {
		return err
	}
	for _, subpath := range subpaths {
		if subpath == fc.getManifestPath() {
			continue
		}
		_, ok := validPaths[subpath]
		if ok {
			continue
		}
		fc.logger.Info("remove untracked path", "path", subpath)
		err = RemovePaths(fc.ctx, subpath)
		if err != nil {
			return err
		}
	}
	return nil
}

// Gets an item from the file cache.
// Returns an error if the get operation fails
func (fc *fileCache) get(key string, dest string) error {
	defer fc.save()
	fc.logger.Info("file cache get", "key", key, "dest", dest)
	item, ok := fc.contents[key]
	if !ok {
		return fmt.Errorf("key not found %s", key)
	}
	unsquashPath := dest
	if item.IsFile {
		unsquashPath = filepath.Dir(unsquashPath)
	}
	_, err := Command(fc.ctx, []string{"unsquashfs", "-force", "-dest", unsquashPath, item.Path}, CmdOpts{}).Run()
	if err != nil {
		return err
	}
	_, err = os.Lstat(dest)
	if err != nil {
		return err
	}
	item.LastAccessed = time.Now()
	item.LastUuid = fc.uuid
	fc.contents[key] = item
	return err
}

// Returns the path to the manifest JSON file
func (fc *fileCache) getManifestPath() string {
	return filepath.Join(fc.dir, "manifest.json")
}

// Returns a boolean indicating whether the file cache has the given key
func (fc *fileCache) hasKey(key string) bool {
	_, ok := fc.contents[key]
	return ok
}

// Initializes the cache.   Creates the cache directory if it doesn't exist, loads the manifest from disk, and cleans the cache.
// Returns an error if any part of the initializtion process fails.
func (fc *fileCache) initialize() error {
	err := CreateDirs(fc.ctx, fc.dir)
	if err != nil {
		return err
	}
	err = fc.load()
	if err != nil {
		return err
	}
	return fc.clean()
}

// Loads on-disk state into the [fileCache] metadata.
// Returns an error if this process fails.
func (fc *fileCache) load() error {
	fc.contents = Map[string, fileCacheItem]{}
	_, err := os.Lstat(fc.getManifestPath())
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	data := fileCacheManifest{}
	err = UnmarshalFile(fc.ctx, fc.getManifestPath(), &data)
	if err != nil {
		fc.logger.Info("manifest unparseable")
		return RemovePaths(fc.ctx, fc.getManifestPath())
	}
	if data.Version != fileCacheVersion {
		fc.logger.Info("manifest version mismatch", "manifest", data.Version, "current", fileCacheVersion)
		return RemovePaths(fc.ctx, fc.getManifestPath())
	}
	fc.contents = Map[string, fileCacheItem](data.Contents)
	return nil
}

// Pops (removes) an item from the file cache.
// Returns an error if this process fails.
func (fc *fileCache) pop(key string) error {
	fc.logger.Info("file cache pop", "key", key)
	item, ok := fc.contents[key]
	if !ok {
		return nil
	}
	defer fc.save()
	err := RemovePaths(fc.ctx, item.Path)
	if err != nil {
		return err
	}
	delete(fc.contents, key)
	return nil
}

// Puts an item (by key) into the cache.
// Returns an error if the put operation fails.
func (fc *fileCache) put(key string, dest string, fetchCb fileCacheFetchCb) error {
	defer fc.save()
	fc.logger.Info("file cache put", "key", key, "dest", dest)
	err := fetchCb(dest)
	if err != nil {
		return err
	}
	lstat, err := os.Lstat(dest)
	if err != nil {
		return err
	}
	err = fc.trim(int(lstat.Size()))
	if err != nil {
		return err
	}
	path := filepath.Join(fc.dir, fmt.Sprintf("%s.squashfs", key))
	isFile := !lstat.IsDir()
	_, err = Command(fc.ctx, []string{"mksquashfs", dest, path}, CmdOpts{}).Run()
	if err != nil {
		return err
	}
	lstat, err = os.Lstat(path)
	if err != nil {
		return err
	}
	fc.contents[key] = fileCacheItem{
		IsFile:       isFile,
		Key:          key,
		LastAccessed: time.Now(),
		LastUuid:     fc.uuid,
		Path:         path,
		Size:         int(lstat.Size()),
	}

	return nil
}

// Sorting function that puts cache items with a current uuid at the end of the list, otherwise sorts by access times.
// This prevents the cache from trimming content that's been accessed during the current session.
func (fc *fileCache) itemSortFunc(a fileCacheItem, b fileCacheItem) int {
	if a.LastUuid != b.LastUuid {
		if a.LastUuid == fc.uuid {
			return 1
		}
		if b.LastUuid == fc.uuid {
			return -1
		}
	}
	return int(a.LastAccessed.Sub(b.LastAccessed).Seconds())
}

// Returns the current cache size by summing the sizes of the cache's contents.
func (fc *fileCache) getCacheSize() int {
	size := 0
	for _, item := range fc.contents {
		size += item.Size
	}
	return size
}

// Trims the cache to its configured size limit.  If an offset is provided, trims to (size limit + offset).
// Returns an error if the trim operation fails.
func (fc *fileCache) trim(offset int) error {
	if fc.sizeLimit == 0 {
		return nil
	}
	sizeLimit := fc.sizeLimit - offset
	if sizeLimit < 0 {
		return fmt.Errorf("size limit %d and offset %d results in negative number", fc.sizeLimit, offset)
	}
	size := fc.getCacheSize()
	if size < sizeLimit {
		return nil
	}
	items := fc.contents.Values()
	slices.SortFunc(items, fc.itemSortFunc)
	for size > sizeLimit && len(items) > 0 {
		fc.logger.Info("trim cache iteration", "count", len(items), "limit", sizeLimit, "size", size)
		item := items[0]
		items = items[1:]
		if item.LastUuid == fc.uuid {
			fc.logger.Info("stopping search", "key", item.Key, "reason", "current uuid")
			break
		}
		err := fc.pop(item.Key)
		if err != nil {
			return err
		}
		size -= item.Size
	}
	if size >= sizeLimit {
		return fmt.Errorf("trim cache failed: %d < %d", size, sizeLimit)
	}
	return nil
}

// Persists the current file cache metadata to the on-disk manifest.
// Returns an error if the file save operation fails.
func (fc *fileCache) save() error {
	data := fileCacheManifest{Contents: fc.contents, Version: fileCacheVersion}
	return MarshalFile(fc.ctx, data, fc.getManifestPath())
}

// Caches a function by key on-disk.
// If the key does not exist, the fetch callback is called
// If the key does exist, the data is fetched from cache.
// Returns an error if any file cache operation fails.
func CacheFile(ctx context.Context, key string, dest string, fetchCb fileCacheFetchCb) error {
	cacheDir, ok := Dirs(ctx)["cache"]
	if !ok {
		Logger(ctx).Info("cache directory unset - bypassing file cache")
		return fetchCb(dest)
	}
	fc := fileCache{ctx: ctx, dir: cacheDir, logger: Logger(ctx), sizeLimit: FileCacheSizeLimit(ctx) * int(math.Pow10(6)), uuid: Uuid(ctx)}
	err := fc.initialize()
	if err != nil {
		return err
	}
	if fc.hasKey(key) {
		return fc.get(key, dest)
	}
	return fc.put(key, dest, fetchCb)
}
