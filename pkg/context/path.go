package context

import (
	"errors"
	"fmt"
	"os"
)

// Creates the provided directories
// Returns an error if any directories fail to create
func (ctx *Context) CreateDirs(paths ...string) error {
	for _, path := range paths {
		_, err := os.Stat(path)
		if errors.Is(err, os.ErrNotExist) {
			ctx.Logger().Info("create directory", "path", path)
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
func (ctx *Context) SetOwnerForPaths(owner User, paths ...string) error {
	err := ctx.CreateDirs(paths...)
	if err != nil {
		return err
	}

	for _, path := range paths {
		ctx.Logger().Info("ensure directory ownership", "owner", owner)
		_, err = ctx.RunCommand([]string{"chown", "-R", fmt.Sprintf("%d:%d", owner.Uid, owner.Gid), path}, CmdOpts{})
		if err != nil {
			return err
		}
	}

	return nil
}

// Creates a symlink from one path to another path.
// Returns an error if the symlink operation fails.
func (ctx *Context) SymlinkDir(from string, to string) error {
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
