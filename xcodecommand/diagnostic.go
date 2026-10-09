package xcodecommand

// This file turns what the parser and the merge decided into messages for the step log.
// Nothing here changes the command line: deleting it, the second return of merge and
// Command.Diagnostics leaves the parser and the merge as they are.

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode"

	shellquote "github.com/kballard/go-shellquote"
)

// DiagnosticKind classifies a finding about a command's additional options.
type DiagnosticKind int

// Diagnostic kinds.
const (
	MalformedOption       DiagnosticKind = iota // unclassifiable argument, passed through; xcodebuild rejects it
	RejectedOption                              // switches xcodebuild's mode or belongs to another action
	ActionInOptions                             // a build action; the command owns its action list
	SuspiciousUserDefault                       // -NAME=value that looks like a build setting; xcodebuild ignores it silently
	RepeatedOption                              // a value option set by the command and again by the user; xcodebuild refuses it
	PreferStepInput                             // a flag that only works with what the step's inputs provide; use the input
	Override                                    // informational: a default or build setting yielded to the user's
	RedundantOption                             // informational: identical to what the command already sets
)

// Diagnostic is one finding about the additional options, for the step log.
type Diagnostic struct {
	Kind    DiagnosticKind
	Message string
}

func (d Diagnostic) String() string {
	return d.Message
}

// Validation says what a constructor does with diagnostics.
type Validation int

// Validation modes.
const (
	// Warn (default) passes every option through and reports; this matches what the
	// steps did before and keeps working setups that trip a check working.
	Warn Validation = iota
	// Fail returns an error for the first non-informational diagnostic.
	Fail
)

func (k DiagnosticKind) informational() bool {
	return k == Override || k == RedundantOption
}

func firstFailure(diagnostics []Diagnostic) error {
	for _, d := range diagnostics {
		if !d.Kind.informational() {
			return fmt.Errorf("invalid additional option: %s", d.Message)
		}
	}
	return nil
}

// lint reports everything about a command's additional options that the user should see.
func lint(user, merged Options, collisions []collision, policy actionPolicy) []Diagnostic {
	var diagnostics []Diagnostic
	diagnostics = append(diagnostics, user.Diagnostics()...)
	diagnostics = append(diagnostics, lintPolicy(user, policy)...)
	diagnostics = append(diagnostics, lintStepInputs(user, collisions)...)
	diagnostics = append(diagnostics, lintShadows(user, merged)...)
	diagnostics = append(diagnostics, lintCollisions(collisions)...)
	return diagnostics
}

// -ENABLE_BITCODE, -MARKETING_VERSION: a build setting typed with a dash, which
// xcodebuild accepts as a user default and silently ignores. Not -IDEFoo.
var upperCaseFlag = regexp.MustCompile(`^-[A-Z][A-Z0-9_]*$`)

// Diagnostics reports what the parser could not read and what xcodebuild would silently
// ignore, each with its fix. A step logs these before its first xcodebuild call.
func (opts Options) Diagnostics() []Diagnostic {
	var diagnostics []Diagnostic
	for _, o := range opts {
		switch {
		case o.Kind == Unknown && o.issue == quotedFlagWithValue:
			diagnostics = append(diagnostics, Diagnostic{Kind: SuspiciousUserDefault, Message: fmt.Sprintf("%q is a flag quoted together with its value. xcodebuild reads it as a user default and ignores it, so it has no effect. Use %s to apply it, or remove it.", o.Name, unquoteFlag(o.Name))})
		case o.Kind == Unknown:
			diagnostics = append(diagnostics, Diagnostic{Kind: MalformedOption, Message: fmt.Sprintf("%q %s", o.Name, malformedMessage(o))})
		case o.Kind == UserDefault && upperCaseFlag.MatchString(o.Name):
			// -ENABLE_BITCODE=NO: the user meant ENABLE_BITCODE=NO
			setting := o.Name[1:] + "=" + shellQuoted(o.Value)
			diagnostics = append(diagnostics, Diagnostic{Kind: SuspiciousUserDefault, Message: fmt.Sprintf("%q looks like the build setting %s with a leading dash. xcodebuild reads it as a user default and ignores it, so it has no effect. Use %s to apply it, or remove it.", o, setting, setting)})
		}
	}
	return diagnostics
}

func malformedMessage(o Option) string {
	switch o.issue {
	case emptyArgument:
		return "is empty. xcodebuild treats it as an unknown build action. Remove it."
	case quotedFlag:
		return fmt.Sprintf("is a flag quoted together with its value. xcodebuild refuses it. Use %s instead.", unquoteFlag(o.Name))
	case invalidFlag:
		return "is not a valid flag. xcodebuild refuses it. Remove it."
	case colonWithoutValue:
		return "has no value after the colon. xcodebuild refuses it. Add the value, or remove it."
	case missingValue:
		return "has no value. xcodebuild refuses it. Add one, or remove it."
	default:
		return `is not a flag, a NAME=value build setting or a build action. xcodebuild treats it as an unknown build action. Quote a value with spaces, for example CODE_SIGN_IDENTITY="Apple Distribution", or remove it.`
	}
}

// lintPolicy reports the options the command refuses; build actions always are.
func lintPolicy(user Options, policy actionPolicy) []Diagnostic {
	var diagnostics []Diagnostic
	for _, o := range user {
		switch o.Kind {
		case Action:
			diagnostics = append(diagnostics, Diagnostic{Kind: ActionInOptions, Message: fmt.Sprintf("%q is a build action. The Step sets the actions itself. Remove it.", o.Name)})
		default:
			if r, rejected := policy.rejects(o); rejected {
				diagnostics = append(diagnostics, Diagnostic{Kind: RejectedOption, Message: fmt.Sprintf("%q does not work with %s. %s Remove it.", o.String(), policy.name, r.reason)})
			}
		}
	}
	return diagnostics
}

// stepInput is a flag a Step input sets: passed in the additional options instead, it is
// reported so the user reaches for the input.
type stepInput struct {
	hint      string
	satisfied []string // value flags that, all set by the user too, make the flag work without the input
}

var stepInputFlags = map[string]stepInput{
	"-allowProvisioningUpdates": {
		hint:      "cannot update provisioning profiles on its own. xcodebuild needs the App Store Connect API key flags, which the Step adds when its automatic code signing is enabled. Remove it and enable automatic code signing instead.",
		satisfied: []string{"-authenticationKeyPath", "-authenticationKeyID", "-authenticationKeyIssuerID"},
	},
}

// lintStepInputs reports user options from stepInputFlags that the command did not derive
// itself. When it did, they collide, and lintCollisions reports them instead.
func lintStepInputs(user Options, collisions []collision) []Diagnostic {
	var diagnostics []Diagnostic
	for _, o := range user {
		input, ok := stepInputFlags[o.Key()]
		if !ok || slices.ContainsFunc(collisions, func(c collision) bool { return c.derived.Key() == o.Key() }) ||
			setsAll(user, input.satisfied) {
			continue
		}
		diagnostics = append(diagnostics, Diagnostic{Kind: PreferStepInput, Message: fmt.Sprintf("%q %s", o, input.hint)})
	}
	return diagnostics
}

// setsAll says whether opts gives every one of the value flags a value. It is false for
// none: a stepInput without satisfied flags always needs its input.
func setsAll(opts Options, flags []string) bool {
	if len(flags) == 0 {
		return false
	}
	for _, flag := range flags {
		if !slices.ContainsFunc(opts, func(o Option) bool { return o.Kind == ValueOption && o.Name == flag && o.Value != "" }) {
			return false
		}
	}
	return true
}

// lintShadows reports a user "-flag=value" whose flag takes a value or is a switch. It is
// not a collision (a user default keys as "-flag="), but xcodebuild ignores the "=" form,
// so the value never applies. The message names what the build uses instead: the merged
// options' -flag, the Step's or the user's, or nothing. A switch is valuedSwitch's.
func lintShadows(user, merged Options) []Diagnostic {
	var diagnostics []Diagnostic
	for _, u := range user {
		if u.Kind != UserDefault {
			continue
		}
		// the last copy is the one that counts; the merge put the user's after the Step's
		var applied Option
		found := false
		for i := len(merged) - 1; i >= 0 && !found; i-- {
			if o := merged[i]; o.Name == u.Name && (o.Kind == ValueOption || o.Kind == Switch) {
				applied, found = o, true
			}
		}
		yours := found && slices.Contains(user, applied)
		if found && applied.Kind == Switch {
			diagnostics = append(diagnostics, valuedSwitch(u, applied, !yours))
			continue
		}

		var message string
		switch {
		case found && yours:
			// -destination generic/platform=tvOS -destination=generic/platform=watchOS
			message = fmt.Sprintf("and uses your %q instead. Change that one to %s %s to apply it, or remove it.", applied, u.Name, shellQuoted(u.Value))
		case found:
			message = fmt.Sprintf("and uses the Step's %q instead. Use %s %s to apply it, or remove it.", applied, u.Name, shellQuoted(u.Value))
		case slices.Contains(freeFormValueFlags, u.Name):
			// -destination=generic/platform=iOS with no -destination anywhere
			message = fmt.Sprintf("so it has no effect. Use %s %s to apply it, or remove it.", u.Name, shellQuoted(u.Value))
		default:
			continue
		}
		diagnostics = append(diagnostics, Diagnostic{Kind: SuspiciousUserDefault, Message: fmt.Sprintf("%q is written with \"=\". xcodebuild reads it as a user default and ignores it, %s", u, message)})
	}
	return diagnostics
}

// lintCollisions describes what the merge did with each derived option the user set too.
func lintCollisions(collisions []collision) []Diagnostic {
	var diagnostics []Diagnostic
	for _, c := range collisions {
		switch c.kind {
		case redundant:
			diagnostics = append(diagnostics, Diagnostic{Kind: RedundantOption, Message: fmt.Sprintf("%q is already set by the Step. Remove it.", c.derived)})
		case overridden:
			if valued := c.user.Filter(func(o Option) bool { return o.Kind != Switch }); c.derived.Kind == Switch && len(valued) > 0 {
				// "-allowProvisioningUpdates YES" or "-allowProvisioningUpdates:YES" keys as the switch
				for _, u := range valued {
					diagnostics = append(diagnostics, valuedSwitch(u, c.derived, true))
				}
				continue
			}
			diagnostics = append(diagnostics, Diagnostic{Kind: Override, Message: fmt.Sprintf("%q replaces the Step's default %q.", c.user.join(), c.derived)})
		case repeated:
			diagnostics = append(diagnostics, Diagnostic{Kind: RepeatedOption, Message: fmt.Sprintf("%s is managed by the Step. %q conflicts with its %q: xcodebuild accepts the flag only once. Remove it.", c.derived.Name, c.user.join(), c.derived)})
		}
	}
	return diagnostics
}

// valuedSwitch reports u, a value given to sw, a switch the command line already has. How
// xcodebuild reads the value depends on how it is written.
func valuedSwitch(u, sw Option, stepSets bool) Diagnostic {
	kind, effect := MalformedOption, fmt.Sprintf("xcodebuild reads %q as a build action and fails", u.Value)
	switch u.Kind {
	case UserDefault:
		// -allowProvisioningUpdates=YES
		kind, effect = SuspiciousUserDefault, "xcodebuild reads it as a user default and ignores it"
	case ColonOption:
		// -allowProvisioningUpdates:YES
		effect = "xcodebuild refuses it as an invalid option"
	}
	set := fmt.Sprintf("%s is already set", sw)
	if stepSets {
		set = fmt.Sprintf("The Step already sets %s", sw)
	}
	return Diagnostic{Kind: kind, Message: fmt.Sprintf("%q gives a value to %s, which takes none, so %s. %s. Remove it.", u, sw, effect, set)}
}

// join renders the options as one command line fragment, for messages.
func (opts Options) join() string {
	return strings.Join(opts.Args(), " ")
}

// shellQuoted renders a value as it has to be written in xcodebuild_options: the inverse
// of SplitAdditionalOptions.
func shellQuoted(value string) string {
	return shellquote.Join(value)
}

// unquoteFlag turns "-destination 'generic/platform=iOS'" (one argument) into the form the
// user meant: the flag, then the value quoted on its own.
func unquoteFlag(arg string) string {
	flag, value := arg, ""
	if i := strings.IndexFunc(arg, unicode.IsSpace); i >= 0 {
		// "-sdk\tmacosx" splits like "-sdk macosx": the parser takes any whitespace
		flag, value = arg[:i], arg[i:]
	}
	value = strings.Trim(strings.TrimSpace(value), "'\"")
	return flag + " " + shellQuoted(value)
}
