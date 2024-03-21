package async

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransitionBuilder(t *testing.T) {
	// Create a new TransitionBuilder
	builder := NewTransitionBuilder()

	// Set up sample actions
	onEventAction := func(e *Event) error {
		return nil
	}
	onEnterAction := func(e *Event) error {
		return nil
	}
	onLeaveAction := func(e *Event) error {
		return nil
	}

	// Build a transition using the builder
	transition := builder.
		From("state1").
		To("state2").
		Event("event1").
		OnEvent(onEventAction).
		OnEnter(onEnterAction).
		OnLeave(onLeaveAction).
		Build()[0]

	// Make assertions about the resulting transition
	assert.NotNil(t, transition)
	assert.Equal(t, "state1", transition.From)
	assert.Equal(t, "state2", transition.To)
	assert.Equal(t, "event1", transition.Event)
	assert.NotNil(t, transition.onEvent)
	assert.NotNil(t, transition.onEnter)
	assert.NotNil(t, transition.onLeave)
}

func TestTransitionBuilderDefaults(t *testing.T) {
	// Create a new TransitionBuilder without specifying actions
	builder := NewTransitionBuilder()

	// Build a transition using the builder
	transition := builder.
		From("state1").
		To("state2").
		Event("event1").
		Build()[0]

	// Make assertions about the resulting transition and default actions
	assert.NotNil(t, transition)
	assert.Nil(t, transition.onEvent)
	assert.Nil(t, transition.onEnter)
	assert.Nil(t, transition.onLeave)
}
