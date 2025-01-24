package helperapi

import "context"

// DepotDownloadOpts are options supplied to [DepotDownload]
type DepotDownloadOpts struct {
	Context context.Context
}

// Uses DepotDownloader to download data from steam.
// Returns an error on failure.
func (api *Api) DepotDownload(appId string, depotId string, manifestId string, dir string, opts DepotDownloadOpts) error {
	copts := CmdOpts{Attach: true, Context: opts.Context}
	_, err := api.RunCommand([]string{"DepotDownloader", "-app", appId, "-depot", depotId, "-manifest", manifestId, "-dir", dir}, copts)
	return err
}
