package certificateutil

import "testing"

func TestCertificateInfoModel_kindPredicates(t *testing.T) {
	tests := []struct {
		commonName string
		dev        bool
		dist       bool
		installer  bool
	}{
		{"Apple Development: Jane Doe (TEAM)", true, false, false},
		{"iPhone Developer: Jane Doe (TEAM)", true, false, false},
		{"iOS Developer: Jane Doe (TEAM)", true, false, false},
		{"Apple Distribution: Acme Inc (TEAM)", false, true, false},
		{"iPhone Distribution: Acme Inc (TEAM)", false, true, false},
		{"3rd Party Mac Developer Installer: Acme Inc (TEAM)", false, false, true},
		{"Mac Installer Distribution: Acme Inc (TEAM)", false, false, true},
		{"Some Unrelated Certificate", false, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.commonName, func(t *testing.T) {
			c := CertificateInfoModel{CommonName: tt.commonName}
			if got := c.IsDevelopment(); got != tt.dev {
				t.Errorf("IsDevelopment() = %v, want %v", got, tt.dev)
			}
			if got := c.IsDistribution(); got != tt.dist {
				t.Errorf("IsDistribution() = %v, want %v", got, tt.dist)
			}
			if got := c.IsInstaller(); got != tt.installer {
				t.Errorf("IsInstaller() = %v, want %v", got, tt.installer)
			}
		})
	}
}
