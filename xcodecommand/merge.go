package xcodecommand

// placed is an option on an assembled command line: whose it is, and whether the merge
// left it off the line. Both are the merge's decisions, not the option's own.
type placed struct {
	Option
	user    bool // from xcodebuild_options; the zero value is the Step's own
	dropped bool // left off the command line: the user set it too, and the Step's copy yielded
}

// merged is an assembled command line: the Step's options, then the user's, in the order
// given. Its Args leaves the dropped ones out; an Options renders every one of its own.
type merged []placed

// options is the command line's options without their marks, for the rules that read
// options alone.
func (m merged) options() Options {
	opts := make(Options, len(m))
	for i, p := range m {
		opts[i] = p.Option
	}
	return opts
}

func (m merged) Args() []string {
	var args []string
	for _, p := range m {
		if !p.dropped {
			args = append(args, p.args()...)
		}
	}
	return args
}

// merge puts the user's options after the Step's. A Step option the user also set is
// dropped if it is a default, a build setting (xcodebuild takes the last value) or a switch
// (a repeat changes nothing). Any other Step option stays: xcodebuild applies every copy of
// flags like -skip-testing, and refuses a repeated single-valued one like -xcconfig
// ("may only be provided once").
func merge(step, user Options) merged {
	var line merged
	for _, o := range step {
		p := placed{Option: o}
		if user.sets(o) {
			p.dropped = o.fallback || o.Kind == Switch || o.Kind == BuildSetting
		}
		line = append(line, p)
	}
	for _, o := range user {
		line = append(line, placed{Option: o, user: true})
	}
	return line
}
