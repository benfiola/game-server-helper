package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type DownloadCb func(path string) error

func Download(ctx Context, url string, downloadCb DownloadCb) error {
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

	ctx.Logger().Info("download", "url", url, "file", tempFile)
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

	return downloadCb(tempFile)
}
