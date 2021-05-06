package appleauth

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"

	"github.com/bitrise-io/go-utils/log"
)

func fetchPrivateKey(privateKeyURL string) ([]byte, string, error) {
	fileURL, err := url.Parse(privateKeyURL)
	if err != nil {
		return nil, "", err
	}

	key, err := copyOrDownloadFile(fileURL)
	if err != nil {
		return nil, "", err
	}

	return key, getKeyID(fileURL), nil
}

func copyOrDownloadFile(u *url.URL) ([]byte, error) {
	// if file -> copy
	if u.Scheme == "file" {
		b, err := ioutil.ReadFile(u.Path)
		if err != nil {
			return nil, err
		}

		return b, err
	}

	// otherwise download
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Errorf("Failed to close file: %s", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	contentBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %s", err)
	}

	return contentBytes, nil
}

func getKeyID(u *url.URL) string {
	var keyID = "Bitrise" // as default if no ID found in file name

	// get the ID of the key from the file
	if matches := regexp.MustCompile(`AuthKey_(.+)\.p8`).FindStringSubmatch(filepath.Base(u.Path)); len(matches) == 2 {
		keyID = matches[1]
	}

	return keyID
}
