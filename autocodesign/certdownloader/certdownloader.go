// Package certdownloader implements a autocodesign.CertificateProvider which fetches Bitrise hosted Xcode codesigning certificates.
package certdownloader

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bitrise-io/go-steputils/input"
	"github.com/bitrise-io/go-utils/filedownloader"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-xcode/autocodesign"
	"github.com/bitrise-io/go-xcode/certificateutil"
)

// CertificateFileURL contains a p12 file URL and passphrase
type CertificateFileURL struct {
	URL, Passphrase string
}

type downloader struct {
	urls []CertificateFileURL
}

// NewDownloader ...
func NewDownloader(urls []CertificateFileURL) autocodesign.CertificateProvider {
	return downloader{
		urls: urls,
	}
}

func (d downloader) GetCertificates() ([]certificateutil.CertificateInfoModel, error) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	var certInfos []certificateutil.CertificateInfoModel

	for i, p12 := range d.urls {
		log.Debugf("Downloading p12 file number %d from %s", i, p12.URL)

		p12CertInfos, err := downloadPKCS12(httpClient, p12.URL, p12.Passphrase)
		if err != nil {
			return nil, err
		}
		log.Debugf("Codesign identities included:\n%s", certsToString(p12CertInfos))

		certInfos = append(certInfos, p12CertInfos...)
	}

	return certInfos, nil
}

// downloadPKCS12 downloads a pkcs12 format file and parses certificates and matching private keys.
func downloadPKCS12(httpClient *http.Client, certificateURL, passphrase string) ([]certificateutil.CertificateInfoModel, error) {
	contents, err := downloadFile(httpClient, certificateURL)
	if err != nil {
		return nil, err
	} else if contents == nil {
		return nil, fmt.Errorf("certificate (%s) is empty", certificateURL)
	}

	infos, err := certificateutil.CertificatesFromPKCS12Content(contents, passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate (%s), err: %s", certificateURL, err)
	}

	return infos, nil
}

func downloadFile(httpClient *http.Client, src string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	downloader := filedownloader.New(httpClient)
	downloader.WithContext(ctx)
	fileProvider := input.NewFileProvider(downloader)

	return fileProvider.Contents(src)
}

func certsToString(certs []certificateutil.CertificateInfoModel) (s string) {
	for i, cert := range certs {
		s += "- "
		s += cert.String()
		if i < len(certs)-1 {
			s += "\n"
		}
	}
	return
}
