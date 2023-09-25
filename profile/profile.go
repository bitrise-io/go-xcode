package profile

/*
	Functionality of the legacy profileutil package:

	MatchTargetAndProfileEntitlements(targetEntitlements plistutil.PlistData, profileEntitlements plistutil.PlistData, profileType ProfileType)
	String(installedCertificates ...certificateutil.CertificateInfoModel)
	IsXcodeManaged
	CheckValidity
	HasInstalledCertificate
	InstalledProvisioningProfileInfos(profileType ProfileType)
	FindProvisioningProfileInfo(uuid string)
	ProvisioningProfileFromContent(content []byte)
	ProvisioningProfileFromFile(pth string)
	InstalledProvisioningProfiles(profileType ProfileType)
	FindProvisioningProfile(uuid string)
*/

import (
	"io"

	"howett.net/plist"

	"github.com/fullsailor/pkcs7"
)

type Profile struct {
	PKCS7Profile *pkcs7.PKCS7
}

func NewProfile(prof *pkcs7.PKCS7) *Profile {
	return &Profile{PKCS7Profile: prof}
}

func NewProfileFromPKCS7File(reader io.Reader) (*Profile, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	pkcs7Profile, err := pkcs7.Parse(content)
	if err != nil {
		return nil, err
	}

	return NewProfile(pkcs7Profile), nil
}

func (prof Profile) Details() (*Details, error) {
	var details Details
	_, err := plist.Unmarshal(prof.PKCS7Profile.Content, &details)
	if err != nil {
		return nil, err
	}
	return &details, nil
}
