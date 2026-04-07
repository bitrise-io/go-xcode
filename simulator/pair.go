package simulator

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bitrise-io/go-utils/v2/command"
)

type PairList struct {
	Pairs map[string]Pair `json:"pairs"`
}

type Pair struct {
	Watch PairDevice `json:"watch"`
	Phone PairDevice `json:"phone"`
	State string     `json:"state"`
}

type PairDevice struct {
	Name  string `json:"name"`
	UDID  string `json:"udid"`
	State string `json:"state"`
}

func (p Pair) IsInactive() bool {
	return strings.Contains(p.State, "(inactive")
}

func (p Pair) IsUnavailable() bool {
	return strings.Contains(p.State, "(unavailable)")
}

type PairManager struct {
	commandFactory command.Factory
}

func NewPairManager(commandFactory command.Factory) PairManager {
	return PairManager{commandFactory: commandFactory}
}

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

func (m PairManager) CreatePair(watchUDID, phoneUDID string) (string, error) {
	cmd := m.commandFactory.Create("xcrun", []string{"simctl", "pair", watchUDID, phoneUDID}, nil)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return "", fmt.Errorf("create pair: %w\n%s", err, out)
	}

	return strings.TrimSpace(out), nil
}

func (m PairManager) ActivatePair(pairUDID string) error {
	cmd := m.commandFactory.Create("xcrun", []string{"simctl", "pair_activate", pairUDID}, nil)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return fmt.Errorf("activate pair: %w\n%s", err, out)
	}

	return nil
}
