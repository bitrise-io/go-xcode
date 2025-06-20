package certificateutil

import (
	"bufio"
	"crypto/x509"
	"fmt"
	"regexp"
	"strings"

	"github.com/bitrise-io/go-utils/command"
)

// InstalledCodesigningCertificateInfos ...
func InstalledCodesigningCertificateInfos() ([]CertificateInfo, error) {
	certificates, err := installedCodesigningCertificates()
	if err != nil {
		return nil, err
	}

	infos := []CertificateInfo{}
	for _, certificate := range certificates {
		if certificate != nil {
			infos = append(infos, NewCertificateInfo(*certificate, nil))
		}
	}

	return infos, nil
}

// InstalledInstallerCertificateInfos ...
func InstalledInstallerCertificateInfos() ([]CertificateInfo, error) {
	certificates, err := InstalledMacAppStoreCertificates()
	if err != nil {
		return nil, err
	}

	infos := []CertificateInfo{}
	for _, certificate := range certificates {
		if certificate != nil {
			infos = append(infos, NewCertificateInfo(*certificate, nil))
		}
	}

	installerCertificates := FilterCertificateInfoModelsByFilterFunc(infos, func(cert CertificateInfo) bool {
		return strings.Contains(cert.CommonName, "Installer")
	})

	return installerCertificates, nil
}

// installedCodesigningCertificates ...
func installedCodesigningCertificates() ([]*x509.Certificate, error) {
	certificateNames, err := InstalledCodesigningCertificateNames()
	if err != nil {
		return nil, err
	}
	return getInstalledCertificatesByNameSlice(certificateNames)
}

// InstalledMacAppStoreCertificates ...
func InstalledMacAppStoreCertificates() ([]*x509.Certificate, error) {
	certificateNames, err := InstalledMacAppStoreCertificateNames()
	if err != nil {
		return nil, err
	}
	return getInstalledCertificatesByNameSlice(certificateNames)
}

// InstalledCodesigningCertificateNames ...
func InstalledCodesigningCertificateNames() ([]string, error) {
	cmd := command.New("security", "find-identity", "-v", "-p", "codesigning")
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return nil, commandError(cmd.PrintableCommandArgs(), out, err)
	}
	return installedCodesigningCertificateNamesFromOutput(out)
}

// InstalledMacAppStoreCertificateNames ...
func InstalledMacAppStoreCertificateNames() ([]string, error) {
	cmd := command.New("security", "find-identity", "-v", "-p", "macappstore")
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return nil, commandError(cmd.PrintableCommandArgs(), out, err)
	}
	return installedCodesigningCertificateNamesFromOutput(out)
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
			certificate, err := CeritifcateFromPemContent([]byte(normalizedOut))
			if err != nil {
				return nil, err
			}

			certificates = append(certificates, certificate)
		}
	}

	return certificates, nil
}

func installedCodesigningCertificateNamesFromOutput(out string) ([]string, error) {
	pettern := `^[0-9]+\) (?P<hash>.*) "(?P<name>.*)"`
	re := regexp.MustCompile(pettern)

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

func commandError(printableCmd string, cmdOut string, cmdErr error) error {
	return fmt.Errorf("%s failed, out: %s, err: %w", printableCmd, cmdOut, cmdErr)
}
