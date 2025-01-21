package utils

import (
	"errors"
	"fmt"
	"os"
)

func CreateDirectories(ctx Context, paths ...string) error {
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

func SetDirectoryOwner(ctx Context, owner User, paths ...string) error {
	err := CreateDirectories(ctx, paths...)
	if err != nil {
		return err
	}

	for _, path := range paths {
		ctx.Logger().Info("set directory owner", "path", path, "owner", owner)
		_, err = RunCommand(ctx, []string{"chown", "-R", fmt.Sprintf("%d:%d", owner.Uid, owner.Gid), path}, CmdOpts{})
		if err != nil {
			return err
		}
	}

	return nil
}
