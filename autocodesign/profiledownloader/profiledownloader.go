package profiledownloader

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bitrise-io/go-steputils/input"
	"github.com/bitrise-io/go-utils/filedownloader"
	"github.com/bitrise-io/go-xcode/profileutil"
	"github.com/fullsailor/pkcs7"
)

type downloader struct {
	urls   []string
	client *http.Client
}

func New(profileURLs []string, client *http.Client) downloader {
	return downloader{
		urls:   profileURLs,
		client: client,
	}
}

func (d downloader) GetProfiles() ([]*pkcs7.PKCS7, error) {
	var profiles []*pkcs7.PKCS7

	for _, url := range d.urls {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		downloader := filedownloader.NewWithContext(ctx, d.client)
		fileProvider := input.NewFileProvider(downloader)

		contents, err := fileProvider.Contents(url)
		if err != nil {
			return nil, err
		} else if contents == nil {
			return nil, fmt.Errorf("profile (%s) is empty", url)
		}

		parsedProfile, err := profileutil.ProvisioningProfileFromContent(contents)
		if err != nil {
			return nil, fmt.Errorf("invalid pkcs7 file format: %w", err)
		}

		profiles = append(profiles, parsedProfile)
	}

	return profiles, nil
}
