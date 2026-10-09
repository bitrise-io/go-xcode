package xcodecommand

// actionPolicy is a command's policy for additional options; anything not listed passes.
type actionPolicy struct {
	name       string
	defaults   []string // derived keys the user's option replaces
	appendable []string // repeatable keys where derived and user entries both stay
}

var (
	// The build family: a step's -destination is a fallback the user's replaces, -arch
	// is repeatable.
	archivePolicy = actionPolicy{
		name:       ActionArchive,
		defaults:   []string{"-destination"},
		appendable: []string{"-arch"},
	}
	buildPolicy = actionPolicy{
		name:       ActionBuild,
		defaults:   []string{"-destination"},
		appendable: []string{"-arch"},
	}
	analyzePolicy = actionPolicy{
		name:       ActionAnalyze,
		defaults:   []string{"-destination", "-resultBundlePath"},
		appendable: []string{"-arch"},
	}
	// build-for-testing takes the test selection flags, which it bakes into the xctestrun.
	buildForTestingPolicy = actionPolicy{
		name:       ActionBuildForTesting,
		defaults:   []string{"-destination"},
		appendable: []string{"-arch"},
	}
	exportArchivePolicy   = actionPolicy{name: "export archive"}
	resolvePackagesPolicy = actionPolicy{name: "resolve packages"}
)
