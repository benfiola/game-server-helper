package helper

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Creates the provided directories
// Returns an error if any directories fail to create
func CreateDirs(ctx context.Context, paths ...string) error {
	for _, path := range paths {
		_, err := os.Stat(path)
		if errors.Is(err, os.ErrNotExist) {
			Logger(ctx).Info("create directory", "path", path)
			err = os.MkdirAll(path, 0755)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// createTempDirCb is a callback invoked once a temporary directory is created via [CreateTempDir]
type createTempDirCb func(path string) error

// Creates a temporary directory and then invokes a callback with the created path.
// Returns an error if the temporary directory fails to create.
// Returns an error if the callback fails.
func CreateTempDir(ctx context.Context, cb createTempDirCb) error {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}
	Logger(ctx).Info("create temp dir", "path", dir)
	defer os.RemoveAll(dir)
	return cb(dir)
}

// Lists the subpaths in the given directory
// Returns an error if the path is not a directory
func ListDir(ctx context.Context, path string) ([]string, error) {
	fail := func(err error) ([]string, error) {
		return nil, err
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return fail(err)
	}
	paths := []string{}
	for _, entry := range entries {
		paths = append(paths, filepath.Join(path, entry.Name()))
	}
	return paths, nil
}

// Removes the provided paths
// Returns an error if any paths fail to remove
func RemovePaths(ctx context.Context, paths ...string) error {
	for _, path := range paths {
		_, err := os.Stat(path)
		if errors.Is(err, os.ErrNotExist) {
			err = nil
		}
		if err != nil {
			return err
		}
		Logger(ctx).Info("remove path", "path", path)
		err = os.RemoveAll(path)
		if err != nil {
			return err
		}
	}
	return nil
}

// Sets the owner for the given directories
// Returns an error if any 'chown' operation fails
func SetOwnerForPaths(ctx context.Context, owner User, paths ...string) error {
	err := CreateDirs(ctx, paths...)
	if err != nil {
		return err
	}

	for _, path := range paths {
		Logger(ctx).Info("set owner", "owner", owner, "path", path)
		_, err = Command(ctx, []string{"chown", "-R", fmt.Sprintf("%d:%d", owner.Uid, owner.Gid), path}, CmdOpts{}).Run()
		if err != nil {
			return err
		}
	}

	return nil
}

// Creates a symlink from one path to another path.
// Returns an error if the symlink operation fails.
func SymlinkDir(ctx context.Context, from string, to string) error {
	Logger(ctx).Info("create symlink", "from", from, "to", to)
	err := os.MkdirAll(to, 0755)
	if err != nil {
		return err
	}

	err = os.RemoveAll(to)
	if err != nil {
		return err
	}

	err = os.MkdirAll(from, 0755)
	if err != nil {
		return err
	}

	return os.Symlink(from, to)
}
