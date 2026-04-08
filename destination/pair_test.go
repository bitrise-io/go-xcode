package destination

import (
	"testing"

	mockcommand "github.com/bitrise-io/go-xcode/v2/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func createCommand(output string) *mockcommand.Command {
	command := new(mockcommand.Command)
	command.On("PrintableCommandArgs").Return("")
	command.On("Run").Return(nil)
	command.On("RunAndReturnExitCode").Return(0, nil)
	command.On("RunAndReturnTrimmedCombinedOutput").Return(output, nil)

	return command
}

func Test_GivenPairManager_WhenListPairs_ThenParsesOutput(t *testing.T) {
	// Given
	commandFactory := new(mockcommand.CommandFactory)
	manager := NewPairManager(commandFactory)

	pairsJSON := `{
		"pairs": {
			"PAIR-UUID-1": {
				"watch": {"name": "Apple Watch Series 11 (46mm)", "udid": "WATCH-UUID", "state": "Shutdown"},
				"phone": {"name": "iPhone 17 Pro", "udid": "PHONE-UUID", "state": "Shutdown"},
				"state": "(active, disconnected)"
			}
		}
	}`

	parameters := []string{"simctl", "list", "pairs", "-j"}
	commandFactory.On("Create", "xcrun", parameters, mock.Anything).Return(createCommand(pairsJSON))

	// When
	pairList, err := manager.ListPairs()

	// Then
	require.NoError(t, err)
	assert.Len(t, pairList.Pairs, 1)

	pair := pairList.Pairs["PAIR-UUID-1"]
	assert.Equal(t, "PHONE-UUID", pair.Phone.UDID)
	assert.Equal(t, "WATCH-UUID", pair.Watch.UDID)
	assert.False(t, pair.IsInactive())

	commandFactory.AssertCalled(t, "Create", "xcrun", parameters, mock.Anything)
}

func Test_GivenPairManager_WhenListPairsWithInactivePair_ThenIsActiveReturnsFalse(t *testing.T) {
	// Given
	commandFactory := new(mockcommand.CommandFactory)
	manager := NewPairManager(commandFactory)

	pairsJSON := `{
		"pairs": {
			"PAIR-UUID-1": {
				"watch": {"name": "Apple Watch", "udid": "W1", "state": "Shutdown"},
				"phone": {"name": "iPhone", "udid": "P1", "state": "Shutdown"},
				"state": "(inactive, disconnected)"
			}
		}
	}`

	parameters := []string{"simctl", "list", "pairs", "-j"}
	commandFactory.On("Create", "xcrun", parameters, mock.Anything).Return(createCommand(pairsJSON))

	// When
	pairList, err := manager.ListPairs()

	// Then
	require.NoError(t, err)
	pair := pairList.Pairs["PAIR-UUID-1"]
	assert.True(t, pair.IsInactive())
}

func Test_GivenPairManager_WhenListPairsWithUnavailablePair_ThenIsUnavailableReturnsTrue(t *testing.T) {
	// Given
	commandFactory := new(mockcommand.CommandFactory)
	manager := NewPairManager(commandFactory)

	pairsJSON := `{
		"pairs": {
			"PAIR-UUID-1": {
				"watch": {"name": "Apple Watch", "udid": "W1", "state": "Shutdown"},
				"phone": {"name": "iPhone", "udid": "P1", "state": "Shutdown"},
				"state": "(unavailable)"
			}
		}
	}`

	parameters := []string{"simctl", "list", "pairs", "-j"}
	commandFactory.On("Create", "xcrun", parameters, mock.Anything).Return(createCommand(pairsJSON))

	// When
	pairList, err := manager.ListPairs()

	// Then
	require.NoError(t, err)
	pair := pairList.Pairs["PAIR-UUID-1"]
	assert.True(t, pair.IsUnavailable())
}

func Test_GivenPairManager_WhenCreatePair_ThenCallsSimctlPair(t *testing.T) {
	// Given
	commandFactory := new(mockcommand.CommandFactory)
	manager := NewPairManager(commandFactory)

	parameters := []string{"simctl", "pair", "WATCH-UUID", "PHONE-UUID"}
	commandFactory.On("Create", "xcrun", parameters, mock.Anything).Return(createCommand("NEW-PAIR-UUID"))

	// When
	pairID, err := manager.CreatePair("WATCH-UUID", "PHONE-UUID")

	// Then
	require.NoError(t, err)
	assert.Equal(t, "NEW-PAIR-UUID", pairID)

	commandFactory.AssertCalled(t, "Create", "xcrun", parameters, mock.Anything)
}

func Test_GivenPairManager_WhenActivatePair_ThenCallsSimctlPairActivate(t *testing.T) {
	// Given
	commandFactory := new(mockcommand.CommandFactory)
	manager := NewPairManager(commandFactory)

	parameters := []string{"simctl", "pair_activate", "PAIR-UUID"}
	commandFactory.On("Create", "xcrun", parameters, mock.Anything).Return(createCommand(""))

	// When
	err := manager.ActivatePair("PAIR-UUID")

	// Then
	require.NoError(t, err)

	commandFactory.AssertCalled(t, "Create", "xcrun", parameters, mock.Anything)
}
