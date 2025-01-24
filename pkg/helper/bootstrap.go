package helper

import (
	"context"
	"os"

	"github.com/benfiola/game-server-helper/pkg/helperapi"
)

// 'Bootstraps' the entrypoint.
// When run as root, will determine a non-root user, take ownership of necessary directories with this non-root user, and then relaunch the entrypoint as this non-root user.
// When run as non-root, will directly launch the entrypoint as the non-root user.
func (h *Helper) Bootstrap(ctx context.Context, api Api) error {
	api.Logger.Info("bootstrap")

	runAsUser, err := api.GetCurrentUser()
	if err != nil {
		return err
	}
	currentUser := runAsUser

	if currentUser.Uid == 0 {
		runAsUser, err = api.GetEnvUser()
		if err != nil {
			return err
		}

		err = api.UpdateUser("server", runAsUser)
		if err != nil {
			return err
		}

		err = api.SetOwnerForPaths(runAsUser, api.Directories.Values()...)
		if err != nil {
			return err
		}
	}

	executable, err := os.Executable()
	if err != nil {
		return err
	}

	_, err = api.RunCommand([]string{executable, "entrypoint"}, helperapi.CmdOpts{Attach: true, Env: os.Environ(), User: runAsUser})
	return err
}
