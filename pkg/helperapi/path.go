package helperapi

import (
	"errors"
	"fmt"
	"os"
)

// Creates the provided directories
// Returns an error if any directories fail to create
func (api *Api) CreateDirs(paths ...string) error {
	for _, path := range paths {
		_, err := os.Stat(path)
		if errors.Is(err, os.ErrNotExist) {
			api.Logger.Info("create directory", "path", path)
			err = os.MkdirAll(path, 0755)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// Sets the owner for the given directories
// Returns an error if any 'chown' operation fails
func (api *Api) SetOwnerForPaths(owner User, paths ...string) error {
	err := api.CreateDirs(paths...)
	if err != nil {
		return err
	}

	for _, path := range paths {
		api.Logger.Info("set owner", "owner", owner, "path", path)
		_, err = api.RunCommand([]string{"chown", "-R", fmt.Sprintf("%d:%d", owner.Uid, owner.Gid), path}, CmdOpts{})
		if err != nil {
			return err
		}
	}

	return nil
}

// Creates a symlink from one path to another path.
// Returns an error if the symlink operation fails.
func (api *Api) SymlinkDir(from string, to string) error {
	api.Logger.Info("create symlink", "from", from, "to", to)
	err := os.MkdirAll(from, 0755)
	if err != nil {
		return err
	}

	err = os.RemoveAll(from)
	if err != nil {
		return err
	}

	err = os.MkdirAll(to, 0755)
	if err != nil {
		return err
	}

	err = os.Symlink(to, from)
	if err != nil {
		return err
	}

	return nil
}

// createTempDirCb is a callback invoked once a temporary directory is created via [CreateTempDir]
type createTempDirCb func(path string) error

// Creates a temporary directory and then invokes a callback with the created path.
// Returns an error if the temporary directory fails to create.
// Returns an error if the callback fails.
func (api *Api) CreateTempDir(cb createTempDirCb) error {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}
	api.Logger.Info("create temp dir", "path", dir)
	defer os.RemoveAll(dir)

	err = cb(dir)
	if err != nil {
		return err
	}

	return nil
}
