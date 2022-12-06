package errorfinder

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want *nsError
	}{
		{
			name: "Real NSError",
			str:  `Error Domain=IDEProvisioningErrorDomain Code=9 ""ios-simple-objc.app" requires a provisioning profile." UserInfo={IDEDistributionIssueSeverity=3, NSLocalizedDescription="ios-simple-objc.app" requires a provisioning profile., NSLocalizedRecoverySuggestion=Add a profile to the "provisioningProfiles" dictionary in your Export Options property list.}`,
			want: &nsError{
				Description: `"ios-simple-objc.app" requires a provisioning profile.`,
				Suggestion:  `Add a profile to the "provisioningProfiles" dictionary in your Export Options property list.`,
			},
		},
		{
			name: "UserInfo properties order changed",
			str:  `Error Domain=IDEProvisioningErrorDomain Code=9 UserInfo={NSLocalizedRecoverySuggestion=Add a profile to the "provisioningProfiles" dictionary in your Export Options property list., IDEDistributionIssueSeverity=3, NSLocalizedDescription="ios-simple-objc.app" requires a provisioning profile.}`,
			want: &nsError{
				Description: `"ios-simple-objc.app" requires a provisioning profile.`,
				Suggestion:  `Add a profile to the "provisioningProfiles" dictionary in your Export Options property list.`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newNSError(tt.str); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewNSError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findXcodebuildErrors(t *testing.T) {
	tests := []struct {
		name string
		out  string
		want []string
	}{
		{
			name: "Regular error",
			out:  `error: exportArchive: "code-sign-test.app" requires a provisioning profile.`,
			want: []string{`error: exportArchive: "code-sign-test.app" requires a provisioning profile.`},
		},
		{
			name: "Regular error with project path prefix",
			out:  `./steps-xcode-archive/_tmp/code-sign-test.xcodeproj: error: No profile for team 'ASDF8V3WYL' matching 'BitriseBot-Wildcard' found: Xcode couldn't find any provisioning profiles matching 'ASDF8V3WYL/BitriseBot-Wildcard'. Install the profile (by dragging and dropping it onto Xcode's dock item) or select a different one in the Signing & Capabilities tab of the target editor. (in target 'code-sign-test' from project 'code-sign-test')`,
			want: []string{`./steps-xcode-archive/_tmp/code-sign-test.xcodeproj: error: No profile for team 'ASDF8V3WYL' matching 'BitriseBot-Wildcard' found: Xcode couldn't find any provisioning profiles matching 'ASDF8V3WYL/BitriseBot-Wildcard'. Install the profile (by dragging and dropping it onto Xcode's dock item) or select a different one in the Signing & Capabilities tab of the target editor. (in target 'code-sign-test' from project 'code-sign-test')`},
		},
		{
			name: "xcodebuild error",
			out: `xcodebuild: error: Failed to build project code-sign-test with scheme code-sign-test.
        Reason: This scheme builds an embedded Apple Watch app. watchOS 9.0 must be installed in order to archive the scheme
        Recovery suggestion: watchOS 9.0 is not installed. To use with Xcode, first download and install the platform`,
			want: []string{`xcodebuild: error: Failed to build project code-sign-test with scheme code-sign-test.
Reason: This scheme builds an embedded Apple Watch app. watchOS 9.0 must be installed in order to archive the scheme
Recovery suggestion: watchOS 9.0 is not installed. To use with Xcode, first download and install the platform`},
		},
		{
			name: "xcodebuild error and regular error with project path prefix",
			out: `./steps-xcode-archive/_tmp/code-sign-test.xcodeproj: error: No profile for team 'ASDF8V3WYL' matching 'BitriseBot-Wildcard' found: Xcode couldn't find any provisioning profiles matching 'ASDF8V3WYL/BitriseBot-Wildcard'. Install the profile (by dragging and dropping it onto Xcode's dock item) or select a different one in the Signing & Capabilities tab of the target editor. (in target 'code-sign-test' from project 'code-sign-test')

xcodebuild: error: Failed to build project code-sign-test with scheme code-sign-test.
        Reason: This scheme builds an embedded Apple Watch app. watchOS 9.0 must be installed in order to archive the scheme
        Recovery suggestion: watchOS 9.0 is not installed. To use with Xcode, first download and install the platform`,
			want: []string{
				`./steps-xcode-archive/_tmp/code-sign-test.xcodeproj: error: No profile for team 'ASDF8V3WYL' matching 'BitriseBot-Wildcard' found: Xcode couldn't find any provisioning profiles matching 'ASDF8V3WYL/BitriseBot-Wildcard'. Install the profile (by dragging and dropping it onto Xcode's dock item) or select a different one in the Signing & Capabilities tab of the target editor. (in target 'code-sign-test' from project 'code-sign-test')`,
				`xcodebuild: error: Failed to build project code-sign-test with scheme code-sign-test.
Reason: This scheme builds an embedded Apple Watch app. watchOS 9.0 must be installed in order to archive the scheme
Recovery suggestion: watchOS 9.0 is not installed. To use with Xcode, first download and install the platform`},
		},
		{
			name: "NSError",
			out:  `Error Domain=IDEProvisioningErrorDomain Code=9 ""code-sign-test.app" requires a provisioning profile." UserInfo={IDEDistributionIssueSeverity=3, NSLocalizedDescription="code-sign-test.app" requires a provisioning profile., NSLocalizedRecoverySuggestion=Add a profile to the "provisioningProfiles" dictionary in your Export Options property list.}`,
			want: []string{`"code-sign-test.app" requires a provisioning profile. Add a profile to the "provisioningProfiles" dictionary in your Export Options property list.`},
		},
		{
			name: "Regular error and NSError pair",
			out: `error: exportArchive: "share-extension.appex" requires a provisioning profile.

Error Domain=IDEProvisioningErrorDomain Code=9 ""share-extension.appex" requires a provisioning profile." UserInfo={IDEDistributionIssueSeverity=3, NSLocalizedDescription="share-extension.appex" requires a provisioning profile., NSLocalizedRecoverySuggestion=Add a profile to the "provisioningProfiles" dictionary in your Export Options property list.}`,
			want: []string{`"share-extension.appex" requires a provisioning profile. Add a profile to the "provisioningProfiles" dictionary in your Export Options property list.`},
		},
		{
			name: "Extra regular error",
			out: `error: exportArchive: "watchkit-app.app" requires a provisioning profile.

Error Domain=IDEProvisioningErrorDomain Code=9 ""watchkit-app.app" requires a provisioning profile." UserInfo={IDEDistributionIssueSeverity=3, NSLocalizedDescription="watchkit-app.app" requires a provisioning profile., NSLocalizedRecoverySuggestion=Add a profile to the "provisioningProfiles" dictionary in your Export Options property list.}

error: exportArchive: "share-extension.appex" requires a provisioning profile.
`,
			want: []string{
				`"watchkit-app.app" requires a provisioning profile. Add a profile to the "provisioningProfiles" dictionary in your Export Options property list.`,
				`error: exportArchive: "share-extension.appex" requires a provisioning profile.`,
			},
		},
		{
			name: "Regular error with NSError pair and xcodebuild error",
			out: `error: exportArchive: "watchkit-app.app" requires a provisioning profile.

Error Domain=IDEProvisioningErrorDomain Code=9 ""watchkit-app.app" requires a provisioning profile." UserInfo={IDEDistributionIssueSeverity=3, NSLocalizedDescription="watchkit-app.app" requires a provisioning profile., NSLocalizedRecoverySuggestion=Add a profile to the "provisioningProfiles" dictionary in your Export Options property list.}

xcodebuild: error: Failed to build project code-sign-test with scheme code-sign-test.
        Reason: This scheme builds an embedded Apple Watch app. watchOS 9.0 must be installed in order to archive the scheme
        Recovery suggestion: watchOS 9.0 is not installed. To use with Xcode, first download and install the platform.
`,
			want: []string{
				`"watchkit-app.app" requires a provisioning profile. Add a profile to the "provisioningProfiles" dictionary in your Export Options property list.`,
				`xcodebuild: error: Failed to build project code-sign-test with scheme code-sign-test.
Reason: This scheme builds an embedded Apple Watch app. watchOS 9.0 must be installed in order to archive the scheme
Recovery suggestion: watchOS 9.0 is not installed. To use with Xcode, first download and install the platform.`,
			},
		},
		{
			name: "All of the various errors appear in the output",
			out: `Error Domain=IDEProvisioningErrorDomain Code=9 ""ios-simple-objc.app" requires a provisioning profile." UserInfo={IDEDistributionIssueSeverity=3, NSLocalizedDescription="ios-simple-objc.app" requires a provisioning profile., NSLocalizedRecoverySuggestion=Add a profile to the "provisioningProfiles" dictionary in your Export Options property list.}
More build stuff
xcodebuild: error: testing failed
Recovery suggestion: test again
A bunch of nonsense
error: archiving failed
More nonsense
 error: store upload failed`,
			want: []string{
				"error: archiving failed",
				" error: store upload failed",
				"\"ios-simple-objc.app\" requires a provisioning profile. Add a profile to the \"provisioningProfiles\" dictionary in your Export Options property list.",
				"xcodebuild: error: testing failed\nRecovery suggestion: test again",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FindXcodebuildErrors(tt.out); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findXcodebuildErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindXcodebuildErrors(t *testing.T) {
	tests := []struct {
		name           string
		output         string
		wantErrorLines []string
	}{
		{
			name: "Single error",
			output: `Error Domain=IDEProvisioningErrorDomain Code=9 ""ios-simple-objc.app" requires a provisioning profile." UserInfo={IDEDistributionIssueSeverity=3, NSLocalizedDescription="ios-simple-objc.app" requires a provisioning profile., NSLocalizedRecoverySuggestion=Add a profile to the "provisioningProfiles" dictionary in your Export Options property list.}
More build stuff
xcodebuild: error: testing failed
Recovery suggestion: test again
A bunch of nonsense
error: archiving failed
More nonsense
 error: store upload failed`,
			wantErrorLines: []string{
				"error: archiving failed",
				" error: store upload failed",
				"\"ios-simple-objc.app\" requires a provisioning profile. Add a profile to the \"provisioningProfiles\" dictionary in your Export Options property list.",
				"xcodebuild: error: testing failed\nRecovery suggestion: test again",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorLines := FindXcodebuildErrors(tt.output)

			if !reflect.DeepEqual(errorLines, tt.wantErrorLines) {
				t.Errorf("got error lines = %s, want %s", errorLines, tt.wantErrorLines)
			}
		})
	}
}
