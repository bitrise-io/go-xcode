package xcodecommand

import (
	"fmt"
	"path/filepath"
)

// ShowBuildSettingsParams describes an `xcodebuild -showBuildSettings -json` invocation.
// Zero-valued fields are omitted. Set either Target or Scheme; a workspace needs a Scheme.
type ShowBuildSettingsParams struct {
	ProjectPath       string // .xcodeproj, .xcworkspace or a Swift package (no flag; run in its directory)
	Target            string
	Scheme            string
	Configuration     string
	AdditionalOptions Options // the step's xcodebuild_options, from ParseOptions, as given to its main command
}

// ShowBuildSettings renders params into a settings query Command. Its stdout is JSON, one
// entry per target the query covers; buildsettings.Reader runs and reads it.
//
// A step reads build settings as a preparation for its main command and passes it the same
// xcodebuild_options: build settings and package flags change the result. Options that do
// not apply to the query (build actions, test-only and mode-switching flags, anything
// unparseable) are left out, and the Command reports them in one informational line.
func ShowBuildSettings(params ShowBuildSettingsParams) (Command, error) {
	// xcodebuild: "You cannot specify both a scheme and targets.", "You cannot specify targets
	// with a workspace.", "If you specify a workspace then you must also specify a scheme.";
	// with neither on a project it reads the first target, which no caller means.
	switch {
	case params.Target != "" && params.Scheme != "":
		return Command{}, fmt.Errorf("reading the build settings of %s takes a target or a scheme, not both", params.ProjectPath)
	case params.Target == "" && params.Scheme == "":
		return Command{}, fmt.Errorf("reading the build settings of %s needs a target or a scheme", params.ProjectPath)
	case filepath.Ext(filepath.Clean(params.ProjectPath)) == xcworkspaceExtension && params.Scheme == "":
		return Command{}, fmt.Errorf("reading the build settings of workspace %s needs a scheme", params.ProjectPath)
	}
	opts := containerOptions(params.ProjectPath)
	opts = appendValue(opts, "-target", params.Target)
	opts = appendValue(opts, "-scheme", params.Scheme)
	opts = appendValue(opts, "-configuration", params.Configuration)
	opts = append(opts, Option{Kind: Switch, Name: "-showBuildSettings"}, Option{Kind: Switch, Name: "-json"})

	// archive -test-iterations 2 -skipPackagePluginValidation FOO=1 -> -skipPackagePluginValidation FOO=1
	user, diagnostics := leaveOut(showBuildSettingsPolicy, params.AdditionalOptions, "the build settings query")
	return Command{args: merge(opts, user).Args(), diagnostics: diagnostics}, nil
}

// The query sets -json itself: it is what makes the output machine-readable.
var showBuildSettingsPolicy = actionPolicy{
	name:       "the build settings query",
	rejections: []rejection{modeSwitching.except("-showBuildSettings"), testOnly},
}
