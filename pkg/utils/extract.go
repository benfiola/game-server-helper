package utils

import (
	"fmt"
	"strings"
)

func Extract(ctx Context, src string, dest string) error {
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
		return fmt.Errorf("unimplemented extension %s", src)
	}

	_, err := RunCommand(ctx, cmd, CmdOpts{})
	return err
}
