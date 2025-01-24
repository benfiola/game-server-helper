package helperapi

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// downloadCb is a callback invoked with a temporary path pointing to a downloaded file
type downloadCb func(path string) error

// Downloads a url to a temporary file and invokes a callback with the path to the temporary file.
// Returns an error if the download fails.
// Returns an error if the callback returns an error.
func (api *Api) Download(url string, cb downloadCb) error {
	tempDir, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	baseName := filepath.Base(url)
	tempFile := filepath.Join(tempDir, baseName)
	handle, err := os.Create(tempFile)
	if err != nil {
		return err
	}
	defer handle.Close()

	api.Logger.Info("download", "url", url, "file", tempFile)
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s sent non-200 status code: %d", url, response.StatusCode)
	}

	chunkSize := 1024 * 1024
	_, err = io.CopyBuffer(handle, response.Body, make([]byte, chunkSize))
	if err != nil {
		return err
	}

	return cb(tempFile)
}
