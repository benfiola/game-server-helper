package helper

import (
	"errors"
	"fmt"
	"os"
)

// DirectoryMap stores a label to path mapping
type DirectoryMap map[string]string

// Returns a list of paths stored in the directory map
func (dm *DirectoryMap) List() []string {
	list := []string{}
	for _, directory := range *dm {
		list = append(list, directory)
	}
	return list
}

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
		api.Logger.Info("ensure directory ownership", "owner", owner, "path", path)
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
