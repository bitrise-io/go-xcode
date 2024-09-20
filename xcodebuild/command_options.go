package xcodebuild

import "reflect"

const (
	optionFieldTag = "xboption"
)

// CommandOptions represents xcodebuild command options.
// The struct fields are tagged with the option key and the struct fields can be of type string or bool.
// If the field is of type bool and true, the option key will be added to the command arguments.
// If the field is of type string and not empty, the option key and the value will be added to the command arguments.
// Support for new field types might require changes in CommandOptions.optionsAndValues and CommandOptions.cmdArgs methods.
type CommandOptions struct {
	Project                   string `xboption:"project"`
	Workspace                 string `xboption:"workspace"`
	Scheme                    string `xboption:"scheme"`
	Target                    string `xboption:"target"`
	Configuration             string `xboption:"configuration"`
	Destination               string `xboption:"destination"`
	XCConfig                  string `xboption:"xcconfig"`
	ArchivePath               string `xboption:"archivePath"`
	SDK                       string `xboption:"sdk"`
	ResultBundlePath          string `xboption:"resultBundlePath"`
	TestPlan                  string `xboption:"testPlan"`
	AllowProvisioningUpdates  bool   `xboption:"allowProvisioningUpdates"`
	AuthenticationKeyPath     string `xboption:"authenticationKeyPath"`
	AuthenticationKeyID       string `xboption:"authenticationKeyID"`
	AuthenticationKeyIssuerID string `xboption:"authenticationKeyIssuerID"`

	CustomOptions map[string]any
}

func (options CommandOptions) cmdArgs() []string {
	optsKeyValues := options.optionsAndValues()

	for key, value := range options.CustomOptions {
		optsKeyValues[key] = value
	}

	var opts []string
	for key, value := range optsKeyValues {
		switch value := value.(type) {
		case bool:
			if value {
				opts = append(opts, key)
			}
		default:
			opts = append(opts, key, value.(string))
		}
	}

	return opts
}

func (options CommandOptions) optionsAndValues() map[string]any {
	optsKeyValues := map[string]any{}

	rv := reflect.Indirect(reflect.ValueOf(&options))
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		tag := field.Tag.Get(optionFieldTag)
		if tag == "" {
			continue
		}

		value := rv.FieldByName(field.Name).Interface()
		if value == nil || isZeroOfUnderlyingType(value) {
			continue
		}

		optsKey := "-" + tag
		optsKeyValues[optsKey] = value
	}

	return optsKeyValues
}

func isZeroOfUnderlyingType(x interface{}) bool {
	return x == reflect.Zero(reflect.TypeOf(x)).Interface()
}
