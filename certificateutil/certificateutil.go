package certificateutil

import (
	"bufio"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"regexp"
	"strings"

	"github.com/bitrise-io/go-utils/v2/command"
)

type FindIdentityPolicy string

const (
	CodesigningPolicy FindIdentityPolicy = "codesigning"
	MacappstorePolicy FindIdentityPolicy = "macappstore"
)

type KeyChainCertificateLister struct {
	cmdFactory command.Factory
}

func NewKeyChainCertificateLister(cmdFactory command.Factory) KeyChainCertificateLister {
	return KeyChainCertificateLister{cmdFactory: cmdFactory}
}

func (l KeyChainCertificateLister) ListCertificateNames(policy FindIdentityPolicy) ([]string, error) {
	cmd := l.cmdFactory.Create("security", []string{"find-identity", "-v", "-p", string(policy)}, nil)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		// TODO: error handling
		return nil, err
	}
	return installedCertificateNamesFromOutput(out)
}

func (l KeyChainCertificateLister) ListCertificates(policy FindIdentityPolicy) ([]*x509.Certificate, error) {
	certificateNames, err := l.ListCertificateNames(policy)
	if err != nil {
		return nil, err
	}

	var certificates []*x509.Certificate
	for _, certificateName := range certificateNames {
		certs, err := l.getInstalledCertificates(certificateName)
		if err != nil {
			return nil, err
		}
		certificates = append(certificates, certs...)
	}
	return certificates, nil
}

func (l KeyChainCertificateLister) getInstalledCertificates(name string) ([]*x509.Certificate, error) {
	cmd := l.cmdFactory.Create("security", []string{"find-certificate", "-c", name, "-p", "-a"}, nil)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		// TODO: error handling
		return nil, err
	}

	normalizedOuts, err := normalizeFindCertificateOut(out)
	if err != nil {
		return nil, err
	}

	var certificates []*x509.Certificate
	for _, normalizedOut := range normalizedOuts {
		certificate, err := NewCertificateFromPemContent([]byte(normalizedOut))
		if err != nil {
			return nil, err
		}

		certificates = append(certificates, certificate)
	}

	return certificates, nil
}

func installedCertificateNamesFromOutput(out string) ([]string, error) {
	pattern := `^[0-9]+\) (?P<hash>.*) "(?P<name>.*)"`
	re := regexp.MustCompile(pattern)

	certificateNameMap := map[string]bool{}
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if matches := re.FindStringSubmatch(line); len(matches) == 3 {
			name := matches[2]
			certificateNameMap[name] = true
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	var names []string
	for name := range certificateNameMap {
		names = append(names, name)
	}
	return names, nil
}

func normalizeFindCertificateOut(out string) ([]string, error) {
	certificateContents := []string{}
	pattern := `(?s)(-----BEGIN CERTIFICATE-----.*?-----END CERTIFICATE-----)`
	matches := regexp.MustCompile(pattern).FindAllString(out, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no certificates found in: %s", out)
	}

	for _, certificateContent := range matches {
		if !strings.HasPrefix(certificateContent, "\n") {
			certificateContent = "\n" + certificateContent
		}
		if !strings.HasSuffix(certificateContent, "\n") {
			certificateContent = certificateContent + "\n"
		}
		certificateContents = append(certificateContents, certificateContent)
	}

	return certificateContents, nil
}

func NewCertificateFromPemContent(content []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(content)
	if block == nil || block.Bytes == nil || len(block.Bytes) == 0 {
		return nil, fmt.Errorf("failed to parse profile from: %s", string(content))
	}
	return NewCertificateFromDERContent(block.Bytes)
}

func NewCertificateFromDERContent(content []byte) (*x509.Certificate, error) {
	return x509.ParseCertificate(content)
}
