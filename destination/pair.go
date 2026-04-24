package destination

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bitrise-io/go-utils/v2/command"
)

// PairList holds the result of `simctl list pairs`.
type PairList struct {
	Pairs map[string]Pair `json:"pairs"`
}

// Pair represents a watch–phone simulator pairing.
type Pair struct {
	Watch PairDevice `json:"watch"`
	Phone PairDevice `json:"phone"`
	State string     `json:"state"`
}

// PairDevice is one device in a simulator pair.
type PairDevice struct {
	Name  string `json:"name"`
	UDID  string `json:"udid"`
	State string `json:"state"`
}

// IsInactive reports whether the pair is in an inactive state.
func (p Pair) IsInactive() bool {
	return strings.Contains(p.State, "(inactive")
}

// IsUnavailable reports whether the pair is unavailable.
func (p Pair) IsUnavailable() bool {
	return strings.Contains(p.State, "(unavailable)")
}

// PairManager manages watch–phone simulator pairings via simctl.
type PairManager struct {
	commandFactory command.Factory
}

// NewPairManager creates a PairManager using the provided command factory.
func NewPairManager(commandFactory command.Factory) PairManager {
	return PairManager{commandFactory: commandFactory}
}

// ListPairs returns all simulator pairs reported by simctl.
func (m PairManager) ListPairs() (PairList, error) {
	cmd := m.commandFactory.Create("xcrun", []string{"simctl", "list", "pairs", "-j"}, nil)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return PairList{}, fmt.Errorf("list pairs: %w\n%s", err, out)
	}

	var list PairList
	if err := json.Unmarshal([]byte(out), &list); err != nil {
		return PairList{}, fmt.Errorf("parse pair list: %w", err)
	}

	return list, nil
}

// CreatePair pairs a watch simulator with a phone simulator and returns the new pair UDID.
func (m PairManager) CreatePair(watchUDID, phoneUDID string) (string, error) {
	cmd := m.commandFactory.Create("xcrun", []string{"simctl", "pair", watchUDID, phoneUDID}, nil)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return "", fmt.Errorf("create pair: %w\n%s", err, out)
	}

	return strings.TrimSpace(out), nil
}

// ActivatePair activates an existing simulator pair.
func (m PairManager) ActivatePair(pairUDID string) error {
	cmd := m.commandFactory.Create("xcrun", []string{"simctl", "pair_activate", pairUDID}, nil)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return fmt.Errorf("activate pair: %w\n%s", err, out)
	}

	return nil
}

// Unpair removes an existing simulator pair.
func (m PairManager) Unpair(pairUDID string) error {
	cmd := m.commandFactory.Create("xcrun", []string{"simctl", "unpair", pairUDID}, nil)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return fmt.Errorf("delete pair: %w\n%s", err, out)
	}

	return nil
}
