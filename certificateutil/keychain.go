package certificateutil

import (
	"bufio"
	"crypto/x509"
	"fmt"
	"regexp"
	"strings"

	"github.com/bitrise-io/go-utils/v2/command"
)

// FindIdentityPolicy is the value passed to `security find-identity -p`.
type FindIdentityPolicy string

const (
	// CodesigningPolicy lists code signing identities.
	CodesigningPolicy FindIdentityPolicy = "codesigning"
	// MacappstorePolicy lists Mac App Store installer identities.
	MacappstorePolicy FindIdentityPolicy = "macappstore"
)

// CertificateLister lists code signing certificates installed in the macOS Keychain.
type CertificateLister interface {
	ListCertificateNames(policy FindIdentityPolicy) ([]string, error)
	ListCertificates(policy FindIdentityPolicy) ([]*x509.Certificate, error)
	ListCertificateInfos(policy FindIdentityPolicy) ([]CertificateInfoModel, error)
}

type keyChainCertificateLister struct {
	cmdFactory command.Factory
}

// NewKeyChainCertificateLister returns a CertificateLister backed by the macOS `security` tool.
func NewKeyChainCertificateLister(cmdFactory command.Factory) CertificateLister {
	return keyChainCertificateLister{cmdFactory: cmdFactory}
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

	names := []string{}
	for name := range certificateNameMap {
		names = append(names, name)
	}
	return names, nil
}

// ListCertificateNames returns the common names of the installed certificates for the given policy.
func (l keyChainCertificateLister) ListCertificateNames(policy FindIdentityPolicy) ([]string, error) {
	cmd := l.cmdFactory.Create("security", []string{"find-identity", "-v", "-p", string(policy)}, &command.Opts{})
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s failed: %w (output: %s)", cmd.PrintableCommandArgs(), err, out)
	}
	return installedCertificateNamesFromOutput(out)
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

// ListCertificates returns the installed certificates for the given policy.
func (l keyChainCertificateLister) ListCertificates(policy FindIdentityPolicy) ([]*x509.Certificate, error) {
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

func (l keyChainCertificateLister) getInstalledCertificates(name string) ([]*x509.Certificate, error) {
	cmd := l.cmdFactory.Create("security", []string{"find-certificate", "-c", name, "-p", "-a"}, &command.Opts{})
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s failed: %w (output: %s)", cmd.PrintableCommandArgs(), err, out)
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

// ListCertificateInfos returns the installed certificate infos for the given policy.
func (l keyChainCertificateLister) ListCertificateInfos(policy FindIdentityPolicy) ([]CertificateInfoModel, error) {
	certificates, err := l.ListCertificates(policy)
	if err != nil {
		return nil, err
	}

	infos := []CertificateInfoModel{}
	for _, certificate := range certificates {
		if certificate != nil {
			infos = append(infos, NewCertificateInfo(*certificate, nil))
		}
	}
	return infos, nil
}
