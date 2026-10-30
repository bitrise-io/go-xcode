package xcodecommand

// This file turns what the parser decided into messages for the step log. Nothing here
// changes the command line.

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	shellquote "github.com/kballard/go-shellquote"
)

// DiagnosticKind classifies a finding about a command's additional options.
type DiagnosticKind int

// Diagnostic kinds.
const (
	MalformedOption       DiagnosticKind = iota // an argument xcodebuild reads differently than meant: unparsable, or a value flag that took the next flag
	RejectedOption                              // switches xcodebuild's mode or belongs to another action
	ActionInOptions                             // a build action; the command owns its action list
	SuspiciousUserDefault                       // -NAME=value that looks like a build setting; xcodebuild ignores it silently
	RepeatedOption                              // a value option set by the command and again by the user; xcodebuild refuses it
	PreferStepInput                             // a flag that only works with what the step's inputs provide; use the input
	Override                                    // informational: a default or build setting yielded to the user's
	RedundantOption                             // informational: identical to what the command already sets
)

// Informational says whether the kind is a note rather than a problem: the command line
// does what the user meant. A step logs the rest as warnings, or fails on them.
func (k DiagnosticKind) Informational() bool {
	return k == Override || k == RedundantOption
}

// Diagnostic is one finding about the additional options, for the step log.
type Diagnostic struct {
	Kind    DiagnosticKind
	Option  Option // the user's option the finding is about
	Message string
}

func (d Diagnostic) String() string {
	return d.Message
}

// report is a Diagnostic about o with a formatted message.
func report(kind DiagnosticKind, o Option, format string, args ...any) Diagnostic {
	return Diagnostic{Kind: kind, Option: o, Message: fmt.Sprintf(format, args...)}
}

// -ENABLE_BITCODE, -MARKETING_VERSION: a build setting typed with a dash, which
// xcodebuild accepts as a user default and silently ignores. Not -IDEFoo.
var upperCaseFlag = regexp.MustCompile(`^-[A-Z][A-Z0-9_]*$`)

// Diagnostics reports what xcodebuild silently does with an option other than what the user
// meant, each with its fix. An argument xcodebuild refuses gets no line: the build fails on
// it with xcodebuild's own message. A step logs these when it reads its inputs, before its
// first xcodebuild call.
func (opts Options) Diagnostics() []Diagnostic {
	var diagnostics []Diagnostic
	for _, o := range opts {
		switch {
		case o.Kind == UserDefault && strings.IndexFunc(o.Name, unicode.IsSpace) >= 0:
			// "-destination generic/platform=iOS" as one argument: a user default named
			// "destination generic/platform", which xcodebuild ignores
			diagnostics = append(diagnostics, report(SuspiciousUserDefault, o, "%q is a flag quoted together with its value. xcodebuild reads it as a user default and ignores it, so it has no effect. Use %s to apply it, or remove it.", o, unquoteFlag(o.String())))
		case o.Kind == ValueOption && o.issue == flagAsValue:
			// -derivedDataPath -quiet: xcodebuild takes -quiet as the path and the switch is lost
			diagnostics = append(diagnostics, report(MalformedOption, o, "%q has no value, so xcodebuild takes %q as its value and that flag does not apply. Add a value, or remove it.", o.Name, o.Value))
		case o.Kind == UserDefault && upperCaseFlag.MatchString(o.Name):
			// -ENABLE_BITCODE=NO: the user meant ENABLE_BITCODE=NO
			setting := o.Name[1:] + "=" + shellQuoted(o.Value)
			diagnostics = append(diagnostics, report(SuspiciousUserDefault, o, "%q looks like the build setting %s with a leading dash. xcodebuild reads it as a user default and ignores it, so it has no effect. Use %s to apply it, or remove it.", o, setting, setting))
		}
	}
	return diagnostics
}

// shellQuoted renders a value as it has to be written in xcodebuild_options: the inverse
// of the split in ParseOptions.
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
