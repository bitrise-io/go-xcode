package certificateutil

import (
	"bufio"
	"crypto/x509"
	"fmt"
	"regexp"
	"strings"

	"github.com/bitrise-io/go-utils/v2/command"
)

type SecurityTool struct {
	commandFactory command.Factory
}

func NewSecurityTool(commandFactory command.Factory) SecurityTool {
	return SecurityTool{commandFactory: commandFactory}
}

func (t SecurityTool) InstalledCodesigningCertificateInfos() ([]CertificateInfo, error) {
	certificates, err := t.installedCodesigningCertificates()
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

func (t SecurityTool) InstalledInstallerCertificateInfos() ([]CertificateInfo, error) {
	certificates, err := t.installedMacAppStoreCertificates()
	if err != nil {
		return nil, err
	}

	var infos []CertificateInfo
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

func (t SecurityTool) InstalledCodesigningCertificateNames() ([]string, error) {
	cmd := t.commandFactory.Create("security", []string{"find-identity", "-v", "-p", "codesigning"}, nil)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return nil, err
	}
	return installedCodesigningCertificateNamesFromOutput(out)
}

func (t SecurityTool) InstalledMacAppStoreCertificateNames() ([]string, error) {
	cmd := t.commandFactory.Create("security", []string{"find-identity", "-v", "-p", "macappstore"}, nil)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return nil, err
	}
	return installedCodesigningCertificateNamesFromOutput(out)
}

func (t SecurityTool) installedCodesigningCertificates() ([]*x509.Certificate, error) {
	certificateNames, err := t.InstalledCodesigningCertificateNames()
	if err != nil {
		return nil, err
	}
	return t.getInstalledCertificatesByNameSlice(certificateNames)
}

func (t SecurityTool) installedMacAppStoreCertificates() ([]*x509.Certificate, error) {
	certificateNames, err := t.InstalledMacAppStoreCertificateNames()
	if err != nil {
		return nil, err
	}
	return t.getInstalledCertificatesByNameSlice(certificateNames)
}

func (t SecurityTool) getInstalledCertificatesByNameSlice(certificateNames []string) ([]*x509.Certificate, error) {
	var certificates []*x509.Certificate

	for _, name := range certificateNames {
		cmd := t.commandFactory.Create("security", []string{"find-certificate", "-c", name, "-p", "-a"}, nil)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		if err != nil {
			return nil, err
		}

		normalizedOuts, err := normalizeFindCertificateOut(out)
		if err != nil {
			return nil, err
		}

		for _, normalizedOut := range normalizedOuts {
			certificate, err := CertificateFromPemContent([]byte(normalizedOut))
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
