package async

import (
	"strings"
	"time"
)

// StateAction represents the action for a specific transition.
type Action func(ev *Event) error

// Transition represents a state transition in the FSM.
type Transition struct {
	From    string
	To      string
	Event   string
	Timeout time.Duration
	onEvent Action
	onLeave Action
	onEnter Action
}

// NewTransition creates a new Transition object with the specified parameters.
//
// Parameters:
// - from: the state to transition from.
// - to: the state to transition to.
// - event: the event that triggers the transition.
// - onEvent: the action to perform when the event occurs.
// - onEnter: the action to perform when entering the new state.
// - onLeave: the action to perform when leaving the current state.
//
// Returns:
// - *Transition: the newly created Transition object.
func NewTransition(from, to, event string, onEvent, onEnter, onLeave Action) *Transition {
	t := &Transition{
		From:    from,
		To:      to,
		Event:   event,
		onEnter: onEnter,
		onLeave: onLeave,
		onEvent: onEvent,
	}

	return t
}

// TransitionBuilder is a builder for creating transitions.
type TransitionBuilder struct {
	from    []string
	to      string
	event   string
	onEvent Action
	onEnter Action
	onLeave Action
	timeout time.Duration
}

// NewTransitionBuilder creates a new TransitionBuilder.
func NewTransitionBuilder() *TransitionBuilder {
	return &TransitionBuilder{}
}

// From sets the source state for the transition.
func (builder *TransitionBuilder) From(from string) *TransitionBuilder {
	builder.from = append(builder.from, from)
	return builder
}

func (builder *TransitionBuilder) FromMany(from []string) *TransitionBuilder {
	builder.from = append(builder.from, from...)
	return builder
}

// To sets the target state for the transition.
func (builder *TransitionBuilder) To(to string) *TransitionBuilder {
	builder.to = to
	return builder
}

// Event sets the event trigger for the transition.
func (builder *TransitionBuilder) Event(event string) *TransitionBuilder {
	builder.event = event
	return builder
}

// OnEvent sets the onEvent action for the transition.
func (builder *TransitionBuilder) OnEvent(action Action) *TransitionBuilder {
	builder.onEvent = action
	return builder
}

// OnEnter sets the onEnter action for the transition.
func (builder *TransitionBuilder) OnEnter(action Action) *TransitionBuilder {
	builder.onEnter = action
	return builder
}

// OnLeave sets the onLeave action for the transition.
func (builder *TransitionBuilder) OnLeave(action Action) *TransitionBuilder {
	builder.onLeave = action
	return builder
}

func (builder *TransitionBuilder) Timeout(timeout time.Duration) *TransitionBuilder {
	builder.timeout = timeout
	return builder
}

// Build creates a Transition instance based on the builder's settings.
func (builder *TransitionBuilder) Build() []*Transition {
	res := make([]*Transition, 0, len(builder.from))

	for i := 0; i < len(builder.from); i++ {
		res = append(res, NewTransition(builder.from[i], builder.to,
			builder.event, builder.onEvent,
			builder.onEnter, builder.onLeave))
	}
	return res
}

// BuildAndAdd builds the Transition and adds it to the FSM instance.
func (builder *TransitionBuilder) BuildAndAdd(fsm *FSM) {
	for i := 0; i < len(builder.from); i++ {
		fsm.AddTransition(builder.Build()[i])
	}
}

type transitionKey string

// makeTransitionKey generates a transition key based on the given event and from state.
//
// Parameters:
// - event: a string representing the event.
// - from: a string representing the from state.
//
// Return type:
// - transitionKey: the generated transition key.
func makeTransitionKey(event string, from string) transitionKey {
	// Preallocate a strings.Builder with enough capacity
	var builder strings.Builder
	builder.Grow(len(event) + len(from) + 1)

	// Join the strings using strings.Builder
	builder.WriteString(event)
	builder.WriteString("_")
	builder.WriteString(from)

	// Get the final joined string
	return transitionKey(builder.String())
}
