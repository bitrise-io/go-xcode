package appstoreconnectclient

import (
	"crypto/x509"
	"fmt"
	"math/big"

	"github.com/bitrise-steplib/steps-ios-auto-provision-appstoreconnect/devportal"

	"github.com/bitrise-io/go-xcode/certificateutil"

	"github.com/bitrise-steplib/steps-ios-auto-provision-appstoreconnect/appstoreconnect"
)

// APICertificateSource ...
type APICertificateSource struct {
	client *appstoreconnect.Client
}

// NewAPICertificateSource ...
func NewAPICertificateSource(client *appstoreconnect.Client) devportal.CertificateSource {
	return &APICertificateSource{
		client: client,
	}
}

// QueryCertificateBySerial ...
func (s *APICertificateSource) QueryCertificateBySerial(serial *big.Int) (devportal.Certificate, error) {
	response, err := s.client.Provisioning.FetchCertificate(serial.Text(16))
	if err != nil {
		return devportal.Certificate{}, err
	}

	certs, err := parseCertificatesResponse([]appstoreconnect.Certificate{response})
	if err != nil {
		return devportal.Certificate{}, err
	}
	return certs[0], nil
}

// QueryAllIOSCertificates returns all iOS certificates from App Store Connect API
func (s *APICertificateSource) QueryAllIOSCertificates() (map[appstoreconnect.CertificateType][]devportal.Certificate, error) {
	typeToCertificates := map[appstoreconnect.CertificateType][]devportal.Certificate{}

	for _, certType := range []appstoreconnect.CertificateType{appstoreconnect.Development, appstoreconnect.IOSDevelopment, appstoreconnect.Distribution, appstoreconnect.IOSDistribution} {
		certs, err := queryCertificatesByType(s.client, certType)
		if err != nil {
			return map[appstoreconnect.CertificateType][]devportal.Certificate{}, err
		}
		typeToCertificates[certType] = certs
	}

	return typeToCertificates, nil
}

func parseCertificatesResponse(response []appstoreconnect.Certificate) ([]devportal.Certificate, error) {
	var certifacteInfos []devportal.Certificate
	for _, resp := range response {
		if resp.Type == "certificates" {
			cert, err := x509.ParseCertificate(resp.Attributes.CertificateContent)
			if err != nil {
				return nil, fmt.Errorf("failed to parse certificate: %s", err)
			}

			certInfo := certificateutil.NewCertificateInfo(*cert, nil)

			certifacteInfos = append(certifacteInfos, devportal.Certificate{
				Certificate: certInfo,
				ID:          resp.ID,
			})
		}
	}

	return certifacteInfos, nil
}

func queryCertificatesByType(client *appstoreconnect.Client, certificateType appstoreconnect.CertificateType) ([]devportal.Certificate, error) {
	nextPageURL := ""
	var certificates []appstoreconnect.Certificate
	for {
		response, err := client.Provisioning.ListCertificates(&appstoreconnect.ListCertificatesOptions{
			PagingOptions: appstoreconnect.PagingOptions{
				Limit: 20,
				Next:  nextPageURL,
			},
			FilterCertificateType: certificateType,
		})
		if err != nil {
			return nil, err
		}
		certificates = append(certificates, response.Data...)

		nextPageURL = response.Links.Next
		if nextPageURL == "" {
			return parseCertificatesResponse(certificates)
		}
	}
}
