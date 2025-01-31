package helper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Downloads a url to the target path
// Returns an error if the download fails.
func Download(ctx context.Context, url string, dest string) error {
	handle, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer handle.Close()

	Logger(ctx).Info("download", "url", url, "file", dest)
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
	return err
}
