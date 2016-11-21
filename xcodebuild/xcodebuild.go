package xcodebuild

import "github.com/bitrise-io/go-utils/cmdex"

const (
	xcodeBuildToolName = "xcodebuild"
)

// CommandModel ...
type CommandModel interface {
	PrintableCmd() string
	Command() *cmdex.CommandModel
}
