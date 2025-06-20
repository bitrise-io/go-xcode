package exportoptionsgenerator

import (
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-xcode/v2/certificateutil"
	"github.com/bitrise-io/go-xcode/v2/timeutil"
)

// CodesignIdentityProvider can list certificate infos.
type CodesignIdentityProvider interface {
	ListCodesignIdentities() ([]certificateutil.CertificateInfo, error)
}

// LocalCodesignIdentityProvider ...
type LocalCodesignIdentityProvider struct {
	commandFactory command.Factory
	timeProvider   timeutil.TimeProvider
}

func NewLocalCodesignIdentityProvider(commandFactory command.Factory, timeProvider timeutil.TimeProvider) LocalCodesignIdentityProvider {
	return LocalCodesignIdentityProvider{
		commandFactory: commandFactory,
		timeProvider:   timeProvider,
	}
}

// ListCodesignIdentities ...
func (p LocalCodesignIdentityProvider) ListCodesignIdentities() ([]certificateutil.CertificateInfo, error) {
	securityTool := certificateutil.NewSecurityTool(p.commandFactory)
	certs, err := securityTool.InstalledCodesigningCertificateInfos()
	if err != nil {
		return nil, err
	}
	certInfo := certificateutil.FilterValidCertificateInfos(certs, p.timeProvider)
	return append(certInfo.ValidCertificates, certInfo.DuplicatedCertificates...), nil
}
