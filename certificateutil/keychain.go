package certificateutil

import (
	"bufio"
	"crypto/x509"
	"fmt"
	"regexp"
	"strings"

	"github.com/bitrise-io/go-utils/command"
)

// FindIdentityPolicy is the value passed to `security find-identity -p`.
type FindIdentityPolicy string

const (
	// CodesigningPolicy lists code signing identities.
	CodesigningPolicy FindIdentityPolicy = "codesigning"
	// MacappstorePolicy lists Mac App Store installer identities.
	MacappstorePolicy FindIdentityPolicy = "macappstore"
)

func commandError(printableCmd string, cmdOut string, cmdErr error) error {
	return fmt.Errorf("%s failed, out: %s: %w", printableCmd, cmdOut, cmdErr)
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

// InstalledCertificateNames returns the common names of the installed certificates for the given policy.
func InstalledCertificateNames(policy FindIdentityPolicy) ([]string, error) {
	cmd := command.New("security", "find-identity", "-v", "-p", string(policy))
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return nil, commandError(cmd.PrintableCommandArgs(), out, err)
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

// InstalledCertificates returns the installed certificates for the given policy.
func InstalledCertificates(policy FindIdentityPolicy) ([]*x509.Certificate, error) {
	certificateNames, err := InstalledCertificateNames(policy)
	if err != nil {
		return nil, err
	}
	return getInstalledCertificatesByNameSlice(certificateNames)
}

func getInstalledCertificatesByNameSlice(certificateNames []string) ([]*x509.Certificate, error) {
	certificates := []*x509.Certificate{}
	for _, name := range certificateNames {
		cmd := command.New("security", "find-certificate", "-c", name, "-p", "-a")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		if err != nil {
			return nil, commandError(cmd.PrintableCommandArgs(), out, err)
		}

		normalizedOuts, err := normalizeFindCertificateOut(out)
		if err != nil {
			return nil, err
		}

		for _, normalizedOut := range normalizedOuts {
			certificate, err := NewCertificateFromPemContent([]byte(normalizedOut))
			if err != nil {
				return nil, err
			}

			certificates = append(certificates, certificate)
		}
	}

	return certificates, nil
}
