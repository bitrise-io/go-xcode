package spaceship

import (
	"bytes"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bitrise-steplib/steps-ios-auto-provision-appstoreconnect/devportal"

	"github.com/bitrise-io/go-steputils/command/gems"
	"github.com/bitrise-io/go-steputils/command/rubycommand"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-xcode/appleauth"
)

//go:embed spaceship
var spaceship embed.FS

// Client ...
type Client struct {
	workDir    string
	authConfig appleauth.AppleID
	teamID     string
}

// NewClient ...
func NewClient(authConfig appleauth.AppleID, teamID string) (*Client, error) {
	dir, err := prepareSpaceship()
	if err != nil {
		return nil, err
	}

	return &Client{
		workDir:    dir,
		authConfig: authConfig,
		teamID:     teamID,
	}, nil
}

// NewSpaceshipDevportalClient ...
func NewSpaceshipDevportalClient(client *Client) devportal.Client {
	return devportal.Client{
		CertificateSource: NewSpaceshipCertificateSource(client),
		DeviceClient:      NewDeviceClient(client),
		ProfileClient:     NewSpaceshipProfileClient(client),
	}
}

type spaceshipCommand struct {
	command              *command.Model
	printableCommandArgs string
}

func (c *Client) createRequestCommand(subCommand string, opts ...string) (spaceshipCommand, error) {
	authParams := []string{
		"--username", c.authConfig.Username,
		"--password", c.authConfig.Password,
		"--session", base64.StdEncoding.EncodeToString([]byte(c.authConfig.Session)),
		"--team-id", c.teamID,
	}
	s := []string{"bundle", "exec", "ruby", "main.rb",
		"--subcommand", subCommand,
	}
	s = append(s, opts...)
	printableCommand := strings.Join(s, " ")
	s = append(s, authParams...)

	spaceshipCmd, err := rubycommand.NewFromSlice(s)
	if err != nil {
		return spaceshipCommand{}, err
	}
	spaceshipCmd.SetDir(c.workDir)

	return spaceshipCommand{
		command:              spaceshipCmd,
		printableCommandArgs: printableCommand,
	}, nil
}

func runSpaceshipCommand(cmd spaceshipCommand) (string, error) {
	var output bytes.Buffer
	outWriter := &output
	cmd.command.SetStdout(outWriter)
	cmd.command.SetStderr(outWriter)

	log.Debugf("$ %s", cmd.printableCommandArgs)
	if err := cmd.command.Run(); err != nil {
		return "", fmt.Errorf("spaceship command failed, output: %s, error: %v", output.String(), err)
	}

	jsonRegexp := regexp.MustCompile(`(?m)^\{.*\}$`)
	match := jsonRegexp.FindString(output.String())
	if match == "" {
		return "", fmt.Errorf("output does not contain response: %s", output.String())
	}

	var response struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal([]byte(match), &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v (%s)", err, match)
	}

	if response.Error != "" {
		return "", fmt.Errorf("failed to query Developer Portal: %s", response.Error)
	}

	return match, nil
}

func prepareSpaceship() (string, error) {
	targetDir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
	}

	fsys, err := fs.Sub(spaceship, "spaceship")
	if err != nil {
		return "", err
	}

	if err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Warnf("%s", err)
			return err
		}

		if d.IsDir() {
			return os.MkdirAll(filepath.Join(targetDir, path), 0700)
		}

		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		if err := os.WriteFile(filepath.Join(targetDir, path), content, 0700); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return "", err
	}

	bundler := gems.Version{Found: true, Version: "2.2.24"}
	installBundlerCommand := gems.InstallBundlerCommand(bundler)
	installBundlerCommand.SetStdout(os.Stdout).SetStderr(os.Stderr)
	installBundlerCommand.SetDir(targetDir)

	fmt.Println()
	log.Donef("$ %s", installBundlerCommand.PrintableCommandArgs())
	if err := installBundlerCommand.Run(); err != nil {
		return "", fmt.Errorf("command failed, error: %s", err)
	}

	fmt.Println()
	cmd, err := gems.BundleInstallCommand(bundler)
	if err != nil {
		return "", fmt.Errorf("failed to create bundle command model, error: %s", err)
	}
	cmd.SetStdout(os.Stdout).SetStderr(os.Stderr)
	cmd.SetDir(targetDir)

	fmt.Println()
	log.Donef("$ %s", cmd.PrintableCommandArgs())
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("Command failed, error: %s", err)
	}

	return targetDir, nil
}
