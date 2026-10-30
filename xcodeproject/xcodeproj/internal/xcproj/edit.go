package xcproj

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/tailscale/hujson"
)

// Editor changes a project.xcproj in place, so every byte it does not change stays as Xcode
// wrote it.
type Editor struct {
	root           hujson.Value
	configurations []string
}

// NewEditor parses the content of a project.xcproj file for editing.
func NewEditor(content []byte) (*Editor, error) {
	// The parsed tree aliases its input.
	root, err := hujson.Parse(bytes.Clone(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse project.xcproj: %w", err)
	}

	configurations, err := configurationNames(&root)
	if err != nil {
		return nil, err
	}

	return &Editor{root: root, configurations: configurations}, nil
}

// SetTargetBuildSetting sets key for one configuration of a target. key may carry conditions
// other than [config=...], such as CODE_SIGN_IDENTITY[sdk=iphoneos*].
//
// A key that applies to every configuration is split into one key per configuration first,
// which is how Xcode records a setting that differs between configurations.
func (e *Editor) SetTargetBuildSetting(targetName, configuration, key, value string) error {
	if !slices.Contains(e.configurations, configuration) {
		return fmt.Errorf("project has no configuration %s", configuration)
	}

	parsed, ok := parseBuildSettingKey(key)
	if !ok || parsed.hasConfigurationCondition() {
		return fmt.Errorf("invalid build setting key: %s", key)
	}

	settings, err := e.targetBuildSettings(targetName)
	if err != nil {
		return err
	}

	configuredKey := configurationKey(parsed, configuration)
	if existing := memberValue(settings, configuredKey); existing != nil {
		existing.Value = hujson.String(value)
		return nil
	}

	if shared := memberValue(settings, key); shared != nil {
		for _, other := range e.configurations {
			otherKey := configurationKey(parsed, other)
			if other != configuration && memberValue(settings, otherKey) == nil {
				insertMember(settings, otherKey, shared.Value)
			}
		}
		insertMember(settings, configuredKey, hujson.String(value))
		removeMember(settings, memberIndex(settings, key))
		return nil
	}

	insertMember(settings, configuredKey, hujson.String(value))
	return nil
}

// Content returns the edited project.xcproj.
func (e *Editor) Content() []byte {
	return e.root.Pack()
}

func configurationNames(root *hujson.Value) ([]string, error) {
	value := root.Find("/configurations")
	if value == nil {
		return nil, nil
	}

	standard, err := hujson.Standardize(value.Pack())
	if err != nil {
		return nil, fmt.Errorf("failed to parse configurations: %w", err)
	}

	var raw []rawConfiguration
	if err := json.Unmarshal(standard, &raw); err != nil {
		return nil, fmt.Errorf("failed to decode configurations: %w", err)
	}

	names := make([]string, 0, len(raw))
	for _, configuration := range raw {
		names = append(names, configuration.Name)
	}
	return names, nil
}

func (e *Editor) targetBuildSettings(targetName string) (*hujson.Object, error) {
	targets, ok := valueOf[*hujson.Array](e.root.Find("/targets"))
	if !ok {
		return nil, fmt.Errorf("project has no targets")
	}

	for i := range targets.Elements {
		target, ok := targets.Elements[i].Value.(*hujson.Object)
		if !ok {
			continue
		}
		name, ok := valueOf[hujson.Literal](memberValue(target, "name"))
		if !ok || name.String() != targetName {
			continue
		}

		settings, ok := valueOf[*hujson.Object](memberValue(target, "build-settings"))
		if !ok || len(settings.Members) == 0 {
			return nil, fmt.Errorf("target %s has no build settings", targetName)
		}
		return settings, nil
	}

	return nil, fmt.Errorf("project has no target %s", targetName)
}

func valueOf[T hujson.ValueTrimmed](value *hujson.Value) (T, bool) {
	if value == nil {
		var zero T
		return zero, false
	}
	typed, ok := value.Value.(T)
	return typed, ok
}

func memberValue(object *hujson.Object, name string) *hujson.Value {
	if index := memberIndex(object, name); index >= 0 {
		return &object.Members[index].Value
	}
	return nil
}

func memberIndex(object *hujson.Object, name string) int {
	return slices.IndexFunc(object.Members, func(member hujson.ObjectMember) bool {
		return memberName(member) == name
	})
}

func memberName(member hujson.ObjectMember) string {
	literal, ok := member.Name.Value.(hujson.Literal)
	if !ok {
		return ""
	}
	return literal.String()
}

// configurationKey puts [config=...] right after the setting name, the order Xcode writes
// conditions in.
func configurationKey(key buildSettingKey, configuration string) string {
	key.conditions = append([]string{configConditionPrefix + configuration}, key.conditions...)
	return key.String()
}

// insertMember keeps the members sorted, as Xcode writes them, and copies the layout of the
// last member, which unlike the first follows a comma.
func insertMember(object *hujson.Object, name string, value hujson.ValueTrimmed) {
	layout := object.Members[len(object.Members)-1]
	inserted := hujson.ObjectMember{
		Name: hujson.Value{
			BeforeExtra: bytes.Clone(layout.Name.BeforeExtra),
			Value:       hujson.String(name),
			AfterExtra:  bytes.Clone(layout.Name.AfterExtra),
		},
		Value: hujson.Value{
			BeforeExtra: bytes.Clone(layout.Value.BeforeExtra),
			Value:       value,
		},
	}

	index := slices.IndexFunc(object.Members, func(member hujson.ObjectMember) bool {
		return memberName(member) > name
	})
	if index < 0 {
		index = len(object.Members)
		moveTrailingComma(&object.Members[index-1].Value, &inserted.Value)
	}

	object.Members = slices.Insert(object.Members, index, inserted)
}

func removeMember(object *hujson.Object, index int) {
	if last := len(object.Members) - 1; index == last && index > 0 {
		moveTrailingComma(&object.Members[index].Value, &object.Members[index-1].Value)
	}
	object.Members = slices.Delete(object.Members, index, index+1)
}

// moveTrailingComma hands the object's trailing comma to a new last member: hujson writes one
// only while the last value's AfterExtra is non-nil.
func moveTrailingComma(from, to *hujson.Value) {
	to.AfterExtra = from.AfterExtra
	if from.AfterExtra != nil {
		from.AfterExtra = []byte{}
	}
}
