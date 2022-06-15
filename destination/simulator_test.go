package destination

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GivenDestinationIsCorrect_WhenSimulatorIsCreated_ThenItReturnsCorrectSimulator(t *testing.T) {
	// Given
	destination := "platform=iOS Simulator,name=iPhone 8 Plus,OS=latest"

	// When
	simulator, err := NewSimulator(destination)

	// Then
	if assert.NoError(t, err) {
		expectedSimulator := &Simulator{Platform: "iOS Simulator", Name: "iPhone 8 Plus", OS: "latest"}
		assert.Equal(t, expectedSimulator, simulator)
	}
}

func Test_GivenDestinationHasNoOS_WhenSimulatorIsCreated_ThenItReturnsSimulatorWithLatestOS(t *testing.T) {
	// Given
	destination := "platform=iOS Simulator,name=iPhone 8 Plus"

	// When
	simulator, err := NewSimulator(destination)

	// Then
	if assert.NoError(t, err) {
		expectedSimulator := &Simulator{Platform: "iOS Simulator", Name: "iPhone 8 Plus", OS: "latest"}
		assert.Equal(t, expectedSimulator, simulator)
	}
}

func Test_GivenDestinationHasNoPlatform_WhenSimulatorIsCreated_ThenItReturnsAnError(t *testing.T) {
	// Given
	destination := "name=iPhone 8 Plus,OS=latest"

	// When
	simulator, err := NewSimulator(destination)

	// Then
	if assert.Error(t, err) {
		assert.Nil(t, simulator)
	}
}

func Test_GivenDestinationHasNoName_WhenSimulatorIsCreated_ThenItReturnsAnError(t *testing.T) {
	// Given
	destination := "platform=iOS Simulator,OS=latest"

	// When
	simulator, err := NewSimulator(destination)

	// Then
	if assert.Error(t, err) {
		assert.Nil(t, simulator)
	}
}

func Test_GivenDestinationHasInvalidKey_WhenSimulatorIsCreated_ThenItReturnsAnError(t *testing.T) {
	// Given
	destination := "invalid=iOS Simulator,name=iPhone 8 Plus,OS=latest"

	// When
	simulator, err := NewSimulator(destination)

	// Then
	if assert.Error(t, err) {
		assert.Nil(t, simulator)
	}
}

func Test_GivenDestinationHasInvalidFormat_WhenSimulatorIsCreated_ThenItReturnsAnError(t *testing.T) {
	// Given
	destination := "platform:iOS Simulator,name:iPhone 8 Plus,OS:latest"

	// When
	simulator, err := NewSimulator(destination)

	// Then
	if assert.Error(t, err) {
		assert.Nil(t, simulator)
	}
}
