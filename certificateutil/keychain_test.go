package certificateutil

import (
	"encoding/pem"
	"testing"
	"time"

	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-xcode/v2/mocks"
	"github.com/stretchr/testify/require"
)

func TestListCertificateInfos_endToEnd(t *testing.T) {
	const name = "Apple Development: John Doe (TEAM1)"
	const findIdentityOutput = `  1) ABC123 "Apple Development: John Doe (TEAM1)"
     1 valid identities found`

	cert, _, err := GenerateTestCertificate(1, "TEAM1", "Acme Inc", name, time.Now().AddDate(1, 0, 0))
	require.NoError(t, err)
	pemContent := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}))

	identityCmd := new(mocks.Command)
	identityCmd.On("RunAndReturnTrimmedCombinedOutput").Return(findIdentityOutput, nil)

	certCmd := new(mocks.Command)
	certCmd.On("RunAndReturnTrimmedCombinedOutput").Return(pemContent, nil)

	mockFactory := new(mocks.CommandFactory)
	mockFactory.On("Create", "security", []string{"find-identity", "-v", "-p", "codesigning"}, &command.Opts{}).Return(identityCmd)
	mockFactory.On("Create", "security", []string{"find-certificate", "-c", name, "-p", "-a"}, &command.Opts{}).Return(certCmd)

	lister := NewKeyChainCertificateLister(mockFactory)
	infos, err := lister.ListCertificateInfos(CodesigningPolicy)

	require.NoError(t, err)
	require.Len(t, infos, 1)
	require.Equal(t, name, infos[0].CommonName)
	require.Equal(t, cert.SerialNumber.String(), infos[0].Serial)
	mockFactory.AssertExpectations(t)
}
