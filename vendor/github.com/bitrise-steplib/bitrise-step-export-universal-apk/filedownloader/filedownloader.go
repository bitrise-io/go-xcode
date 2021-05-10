package filedownloader

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/bitrise-io/go-utils/log"
)

// HTTPClient ...
type HTTPClient interface {
	Get(source string) (*http.Response, error)
}

// HTTPFileDownloader ...
type HTTPFileDownloader struct {
	client HTTPClient
}

// New ...
func New(client HTTPClient) HTTPFileDownloader {
	return HTTPFileDownloader{client}
}

// GetWithFallback downloads a file from a given source. Provided destination should be a file that does not exist.
// You can specify fallback sources which will be used in order if downloading fails from either source.
func (downloader HTTPFileDownloader) GetWithFallback(destination, source string, fallbackSources ...string) error {
	sources := append([]string{source}, fallbackSources...)
	for _, source := range sources {
		err := downloader.Get(destination, source)
		if err != nil {
			log.Errorf("Could not download file from: %s", err)
		} else {
			log.Infof("URL used to download file: %s", source)
			return nil
		}
	}
	return fmt.Errorf("None of the sources returned 200 OK status")
}

// Get downloads a file from a given source. Provided destination should be a file that does not exist.
func (downloader HTTPFileDownloader) Get(destination, source string) error {

	resp, err := downloader.client.Get(source)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unable to download file from: %s. Status code: %d", source, resp.StatusCode)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Errorf("Failed to close body, error: %s", err)
		}
	}()

	f, err := os.Create(destination)
	if err != nil {
		return err
	}

	defer func() {
		if err := f.Close(); err != nil {
			log.Errorf("Failed to close file, error: %s", err)
		}
	}()

	if _, err = io.Copy(f, resp.Body); err != nil {
		return err
	}

	return nil
}
