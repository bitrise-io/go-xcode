package ruby

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
)

// Runner is the struct to execute a Ruby script.
type Runner struct {
	inputs       map[string]string
	script       string
	gemfilePth   string
	output       interface{}
	dependencies []string
}

// NewRunner creates a new Runner instance.
func NewRunner(script string, inputs map[string]string) *Runner {
	return &Runner{
		script: script,
		inputs: inputs,
	}
}

// BundleInstall installs the given gems.
// gems can be provided as a gem_name: gem_version,
// define empty string for version, to use the latest version of the gem.
func (r *Runner) BundleInstall(gems map[string]string) error {
	tmpDir, err := pathutil.NormalizedOSTempDirPath("Runner")
	if err != nil {
		return err
	}

	gemfilePth, err := createGemfile(tmpDir, gems)
	if err != nil {
		return err
	}

	cmd := command.New("bundle", "install").AppendEnvs("BUNDLE_GEMFILE=" + gemfilePth)
	if out, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		return errors.New(out)
	}

	r.gemfilePth = gemfilePth

	return nil
}

// Execute executes a Ruby script and parses the script output into the given struct.
func (r *Runner) Execute(output interface{}) error {
	tmpDir, err := pathutil.NormalizedOSTempDirPath("Runner")
	if err != nil {
		return err
	}
	scriptPth := filepath.Join(tmpDir, "script.rb")

	if err := fileutil.WriteStringToFile(scriptPth, r.script); err != nil {
		return err
	}

	args := []string{"ruby", scriptPth}
	var envs []string

	for key, value := range r.inputs {
		envs = append(envs, key+"="+value)
	}

	if r.gemfilePth != "" {
		args = append([]string{"bundle", "exec"}, args...)
		envs = append(envs, "BUNDLE_GEMFILE="+r.gemfilePth)
	}

	cmd := command.New(args[0], args[1:]...).AppendEnvs(envs...)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		fmt.Println(err)
	}

	return json.Unmarshal([]byte(out), output)
}
