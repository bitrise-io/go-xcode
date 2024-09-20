package xcodebuild

import "reflect"

const (
	settingFieldTag = "xbsetting"
)

type CommandBuildSettings struct {
	CodeSigningAllowed  string `xbsetting:"CODE_SIGNING_ALLOWED"`
	CustomBuildSettings map[string]string
}

func (buildSettings CommandBuildSettings) toCmdArgs() []string {
	settingKeyValues := buildSettings.keysAndValues()

	for key, value := range buildSettings.CustomBuildSettings {
		settingKeyValues[key] = value
	}

	var opts []string
	for key, value := range settingKeyValues {
		opts = append(opts, key, value)
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
		settingKeyValues[settingKey] = value.(string)
	}

	return settingKeyValues
}
