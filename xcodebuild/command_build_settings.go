package xcodebuild

import "reflect"

const (
	settingFieldTag = "xbsetting"
)

// CommandBuildSettings represents Xcode build settings to be passed to an xcodebuild command.
// The struct fields are tagged with the build setting key and the struct fields can be of type *bool or string.
// If the field is of type *bool and not nil, the value will be converted to "YES" or "NO" accordingly.
// Support for new field types should be implemented in CommandBuildSettings.keysAndValues method.
type CommandBuildSettings struct {
	CodeSigningAllowed           *bool `xbsetting:"CODE_SIGNING_ALLOWED"`
	GCCInstrumentProgramFlowArcs *bool `xbsetting:"GCC_INSTRUMENT_PROGRAM_FLOW_ARCS"`
	GCCGenerateTestCoverageFiles *bool `xbsetting:"GCC_GENERATE_TEST_COVERAGE_FILES"`
	CompilerIndexStoreEnable     *bool `xbsetting:"COMPILER_INDEX_STORE_ENABLE"`
}

func (buildSettings CommandBuildSettings) cmdArgs() []string {
	settingKeyValues := buildSettings.keysAndValues()

	var opts []string
	for key, value := range settingKeyValues {
		opts = append(opts, key+"="+value)
	}

	return opts
}

func (buildSettings CommandBuildSettings) keysAndValues() map[string]string {
	settingKeyValues := map[string]string{}

	rv := reflect.Indirect(reflect.ValueOf(&buildSettings))
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		tag := field.Tag.Get(settingFieldTag)
		if tag == "" {
			continue
		}

		value := rv.FieldByName(field.Name).Interface()
		if value == nil || isZeroOfUnderlyingType(value) {
			continue
		}

		settingKey := tag
		switch value := value.(type) {
		case *bool:
			if *value {
				settingKeyValues[settingKey] = "YES"
			} else {
				settingKeyValues[settingKey] = "NO"
			}
		case string:
			settingKeyValues[settingKey] = value
		}
	}

	return settingKeyValues
}
