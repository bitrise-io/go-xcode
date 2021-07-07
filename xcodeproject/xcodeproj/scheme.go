package xcodeproj

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/bitrise-io/go-xcode/xcodeproject/xcscheme"
)

const (
	yes         = "YES"
	no          = "NO"
	buildableID = "primary"
)

func (p XcodeProj) saveSharedScheme(scheme xcscheme.Scheme) error {
	dir := filepath.Join(p.Path, "xcshareddata", "xcschemes")
	path := filepath.Join(dir, fmt.Sprintf("%s.xcscheme", scheme.Name))

	contents, err := scheme.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal Scheme: %v", err)
	}

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	if err := ioutil.WriteFile(path, contents, 0600); err != nil {
		return fmt.Errorf("failed to write Scheme file (%s): %v", path, err)
	}

	return nil
}

// ReCreateSharedSchemes creates new shared schemes based on Targets
func (p XcodeProj) ReCreateSharedSchemes() error {
	for _, target := range p.Proj.Targets {
		if !target.IsExecutableProduct() {
			continue
		}

		var uiTestTargets []Target
		for _, target := range p.Proj.Targets {
			if target.IsUITestProduct() || target.IsTestProduct() {
				uiTestTargets = append(uiTestTargets, target)
			}
		}

		scheme := newScheme(target, uiTestTargets, filepath.Base(p.Path))
		if err := p.saveSharedScheme(scheme); err != nil {
			return err
		}
	}

	return nil
}

func newScheme(buildTarget Target, testTargets []Target, projectname string) xcscheme.Scheme {
	return xcscheme.Scheme{
		Name: buildTarget.Name,
		//
		LastUpgradeVersion: "1240",
		Version:            "1.3",
		BuildAction:        newBuildAction(buildTarget, projectname),
		ArchiveAction:      newArchiveAction(buildTarget),
		TestAction:         newTestAction(buildTarget, testTargets, projectname),
	}
}

func newBuildableReference(target Target, projectName string) xcscheme.BuildableReference {
	return xcscheme.BuildableReference{
		BuildableIdentifier: buildableID,
		BlueprintIdentifier: target.ID,
		BuildableName:       path.Base(target.ProductReference.Path),
		BlueprintName:       target.Name,
		ReferencedContainer: fmt.Sprintf("container:%s", projectName),
	}
}

func newBuildAction(target Target, projectName string) xcscheme.BuildAction {
	return xcscheme.BuildAction{
		ParallelizeBuildables:     yes,
		BuildImplicitDependencies: yes,
		BuildActionEntries: []xcscheme.BuildActionEntry{
			{
				BuildForTesting:    yes,
				BuildForRunning:    yes,
				BuildForProfiling:  yes,
				BuildForArchiving:  yes,
				BuildForAnalyzing:  yes,
				BuildableReference: newBuildableReference(target, projectName),
			},
		},
	}
}

func newTestableReference(target Target, projectName string) xcscheme.TestableReference {
	return xcscheme.TestableReference{
		Skipped:            no,
		BuildableReference: newBuildableReference(target, projectName),
	}
}

func newTestAction(buildTarget Target, testTargets []Target, projectName string) xcscheme.TestAction {
	if len(testTargets) == 0 {
		return xcscheme.TestAction{}
	}

	testAction := xcscheme.TestAction{
		BuildConfiguration:           testTargets[0].BuildConfigurationList.DefaultConfigurationName,
		SelectedDebuggerIdentifier:   "Xcode.DebuggerFoundation.Debugger.LLDB",
		SelectedLauncherIdentifier:   "Xcode.DebuggerFoundation.Launcher.LLDB",
		ShouldUseLaunchSchemeArgsEnv: yes,
		MacroExpansion: xcscheme.MacroExpansion{
			BuildableReference: newBuildableReference(buildTarget, projectName),
		},
		Testables: []xcscheme.TestableReference{},
	}

	for _, testTarget := range testTargets {
		testAction.Testables = append(
			testAction.Testables,
			newTestableReference(testTarget, projectName),
		)
	}

	return testAction
}

func newArchiveAction(target Target) xcscheme.ArchiveAction {
	return xcscheme.ArchiveAction{
		BuildConfiguration:       target.BuildConfigurationList.DefaultConfigurationName,
		RevealArchiveInOrganizer: yes,
	}
}
