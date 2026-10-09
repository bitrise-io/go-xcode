package xcodecommand

// actionPolicy is a command's policy for additional options; anything not listed passes.
type actionPolicy struct {
	name       string
	defaults   []string // derived keys the user's option replaces
	appendable []string // repeatable keys where derived and user entries both stay
}
