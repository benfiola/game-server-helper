package helper

import (
	"context"
	"fmt"
	"strings"
)

// Extracts a src archive to a dest folder.
// Returns a failure if the archive type is unrecongized.
// Returns a failure if the extract operation fails.
func Extract(ctx context.Context, src string, dest string) error {
	Logger(ctx).Info("extract", "src", src, "dest", dest)

	err := CreateDirs(ctx, dest)
	if err != nil {
		return err
	}

	var cmd []string
	if strings.HasSuffix(src, ".rar") {
		cmd = []string{"unrar", "-f", "-x", src, dest}
	} else if strings.HasSuffix(src, ".tar.gz") {
		cmd = []string{"tar", "--overwrite", "-xzf", src, "-C", dest}
	} else if strings.HasSuffix(src, ".zip") {
		cmd = []string{"unzip", "-o", src, "-d", dest}
	} else if strings.HasSuffix(src, ".7z") {
		cmd = []string{"7z", "x", src, fmt.Sprintf("-o%s", dest)}
	} else {
		return fmt.Errorf("unrecongized file type %s", src)
	}

	_, err = Command(ctx, cmd, CmdOpts{}).Run()
	return err
}
