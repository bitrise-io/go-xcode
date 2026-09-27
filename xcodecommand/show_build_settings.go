package xcodecommand

// ShowBuildSettingsParams describes an `xcodebuild -showBuildSettings -json` invocation.
// Zero-valued fields are omitted; set Target or Scheme, and a workspace needs a Scheme.
type ShowBuildSettingsParams struct {
	ProjectPath       string // .xcodeproj, .xcworkspace or a Swift package (no flag; run in its directory)
	Target            string
	Scheme            string
	Configuration     string
	AdditionalOptions []string // build settings and package flags from the step's xcodebuild_options
	Validation        Validation
}

// ShowBuildSettings renders params into a settings query Command. Its stdout is JSON, one
// entry per target the query covers; buildsettings.Parse reads it.
func ShowBuildSettings(params ShowBuildSettingsParams) (Command, error) {
	opts := containerOptions(params.ProjectPath)
	opts = appendValue(opts, "-target", params.Target)
	opts = appendValue(opts, "-scheme", params.Scheme)
	opts = appendValue(opts, "-configuration", params.Configuration)
	opts = append(opts, Option{Kind: Switch, Name: "-showBuildSettings"}, Option{Kind: Switch, Name: "-json"})

	return assemble(opts, params.AdditionalOptions, showBuildSettingsPolicy, params.Validation)
}
