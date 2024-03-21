package async

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFSM_HandleEvent(t *testing.T) {
	fsm := NewFSM("stateA", nil)

	transition := NewTransition("stateA", "stateB", "eventX", nil, nil, nil)
	fsm.AddTransition(transition)

	err := fsm.HandleEvent("eventX", nil)
	assert.NoError(t, err)

	assert.Equal(t, "stateB", fsm.CurrentState())
}

func TestFSM_CurrentState(t *testing.T) {
	fsm := NewFSM("stateA", nil)
	assert.Equal(t, "stateA", fsm.CurrentState())
}

func TestFSM_SendEvent(t *testing.T) {
	fsm := NewFSM("stateA", nil)
	err := fsm.SendEvent("eventX", nil)
	assert.NoError(t, err)
}

func TestFSM_HandleEventImpl(t *testing.T) {
	fsm := NewFSM("stateA", nil)

	transition := NewTransition("stateA", "stateB", "eventX", nil, nil, nil)
	fsm.AddTransition(transition)

	ei := eventInfo{
		event: "eventX",
		data:  nil,
	}
	err := fsm.handleEventImpl(ei)
	assert.NoError(t, err)
	assert.Equal(t, "stateB", fsm.CurrentState())
}

func TestFSM_StartAndStop(t *testing.T) {
	fsm := NewFSM("stateA", nil)

	transition := NewTransition("stateA", "stateB", "eventX", nil, nil, nil)
	fsm.AddTransition(transition)

	wg := &sync.WaitGroup{}
	fsm.Start(wg)

	err := fsm.SendEvent("eventX", nil)
	assert.NoError(t, err)

	time.Sleep(time.Millisecond) // Allow time for the FSM goroutine to handle the event

	assert.Equal(t, "stateB", fsm.CurrentState())

	fsm.Stop()
	wg.Wait()
}

func Test_NewTransition(t *testing.T) {
	t1 := NewTransition("from", "to", "event", nil, nil, nil)
	assert.Equal(t, "from", t1.From)
	assert.Equal(t, "to", t1.To)
	assert.Equal(t, "event", t1.Event)
	assert.Nil(t, t1.onEvent)
	assert.Nil(t, t1.onEnter)
	assert.Nil(t, t1.onLeave)
}

func Test_GenerateMermaidDiagram(t *testing.T) {
	fsm := NewFSM("stateA", nil)
	onEvent := func(ev *Event) error {
		return ErrNoTransition
	}
	transition := NewTransition("stateA", "stateB", "eventX", onEvent, nil, nil)
	fsm.AddTransition(transition)

	diagram := GenerateMermaidDiagram(fsm, false)
	assert.Contains(t, diagram, "[*] --> stateA")
	assert.Contains(t, diagram, "stateA --> stateB : eventX")

	diagramWithFunc := GenerateMermaidDiagram(fsm, true)
	assert.Contains(t, diagramWithFunc, "stateA --> stateB : eventX")
}

func TestFSM_HandleEventImpl_ValidTransition(t *testing.T) {
	fsm := NewFSM("stateA", nil)

	transition := NewTransition("stateA", "stateB", "eventX", nil, nil, nil)
	fsm.AddTransition(transition)

	ei := eventInfo{
		event: "eventX",
		data:  nil,
	}
	err := fsm.handleEventImpl(ei)
	assert.NoError(t, err)
	assert.Equal(t, "stateB", fsm.CurrentState())
}

func TestFSM_HandleEventImpl_NoTransition(t *testing.T) {
	fsm := NewFSM("stateA", nil)

	ei := eventInfo{
		event: "eventX",
		data:  nil,
	}

	err := fsm.handleEventImpl(ei)
	assert.Error(t, err)
	assert.Equal(t, ErrEventNotSupported, err)
	assert.Equal(t, "stateA", fsm.CurrentState())
}

func TestFSM_HandleEventImpl_ErrorOnEvent(t *testing.T) {
	fsm := NewFSM("stateA", nil)

	transition := NewTransition("stateA", "stateB", "eventX", func(ev *Event) error {
		return fmt.Errorf("error 13")
	}, nil, nil)
	fsm.AddTransition(transition)

	ei := eventInfo{
		event: "eventX",
		data:  nil,
	}
	err := fsm.handleEventImpl(ei)
	assert.Error(t, err)
	assert.Equal(t, "on event eventX from stateA : error 13", err.Error())
	assert.Equal(t, "stateA", fsm.CurrentState())
}

func TestFSM_HandleEventImpl_ErrorOnEnter(t *testing.T) {
	fsm := NewFSM("stateA", nil)

	transition := NewTransition("stateA", "stateB", "eventX", nil, func(ev *Event) error {
		return fmt.Errorf("error 14")
	}, nil)
	fsm.AddTransition(transition)

	ei := eventInfo{
		event: "eventX",
		data:  nil,
	}
	err := fsm.handleEventImpl(ei)
	assert.Error(t, err)
	assert.Equal(t, "on enter error: error 14", err.Error())
	assert.Equal(t, "stateA", fsm.CurrentState())
}

func TestFSM_HandleEventImpl_ErrorOnLeave(t *testing.T) {
	fsm := NewFSM("stateA", nil)

	transition := NewTransition("stateA", "stateB", "eventX", nil, nil, func(ev *Event) error {
		return fmt.Errorf("error 15")
	})
	fsm.AddTransition(transition)

	ei := eventInfo{
		event: "eventX",
		data:  nil,
	}
	err := fsm.handleEventImpl(ei)
	assert.NoError(t, err)
	assert.Equal(t, "stateB", fsm.CurrentState())
}

func TestFSM_HandleEventImpl_NoTransitionOnEventData(t *testing.T) {
	fsm := NewFSM("stateA", nil)

	transition := NewTransition("stateA", "stateB", "eventX", func(ev *Event) error {
		return ErrNoTransition
	}, nil, nil)
	fsm.AddTransition(transition)

	ei := eventInfo{
		event: "eventX",
		data:  "some data",
	}
	err := fsm.handleEventImpl(ei)
	assert.Equal(t, err, ErrNoTransition)
	assert.Equal(t, "stateA", fsm.CurrentState())
}

func TestFSM_HandleEventImpl_BeforeEvent(t *testing.T) {
	fsm := NewFSM("stateA", nil)
	var beforeEventCalled bool

	fsm.BeforeEvent = func(ev *Event) {
		beforeEventCalled = true
	}

	transition := NewTransition("stateA", "stateB", "eventX", nil, nil, nil)
	fsm.AddTransition(transition)

	ei := eventInfo{
		event: "eventX",
		data:  nil,
	}
	err := fsm.handleEventImpl(ei)
	assert.NoError(t, err)
	assert.True(t, beforeEventCalled)
}

func TestFSM_HandleEventImpl_AfterEvent(t *testing.T) {
	fsm := NewFSM("stateA", nil)
	var afterEventCalled bool

	fsm.AfterEvent = func(ev *Event, err error) {
		afterEventCalled = true
	}

	transition := NewTransition("stateA", "stateB", "eventX", nil, nil, nil)
	fsm.AddTransition(transition)

	ei := eventInfo{
		event: "eventX",
		data:  nil,
	}
	err := fsm.handleEventImpl(ei)
	assert.NoError(t, err)
	assert.True(t, afterEventCalled)
}

// Create a test case to cover transition caching
func Test_FSM_HandleEventImpl_Cache(t *testing.T) {
	// Create a new FSM for testing
	fsm := NewFSM("zero", nil)

	eventPlusOne := "plus-one"
	eventMinusOne := "minus-one"

	// Add the transitions to the FSM
	fsm.AddTransition(NewTransition("zero", "one", eventPlusOne, nil, nil, nil))
	fsm.AddTransition(NewTransition("one", "two", eventPlusOne, nil, nil, nil))
	fsm.AddTransition(NewTransition("two", "three", eventPlusOne, nil, nil, nil))
	fsm.AddTransition(NewTransition("three", "two", eventMinusOne, nil, nil, nil))

	// Define an eventInfo for testing
	data := "test data"
	// Call the public handleEvent function multiple times
	callHandleEvent(fsm, eventInfo{event: eventPlusOne, data: data}, t)
	callHandleEvent(fsm, eventInfo{event: eventPlusOne, data: data}, t)
	callHandleEvent(fsm, eventInfo{event: eventPlusOne, data: data}, t)
	callHandleEvent(fsm, eventInfo{event: eventMinusOne, data: data}, t)
	callHandleEvent(fsm, eventInfo{event: eventPlusOne, data: data}, t)
}

func callHandleEvent(fsm *FSM, ei eventInfo, t *testing.T) {
	err := fsm.HandleEvent(ei.event, ei.data)
	if err != nil {
		t.Errorf("Error occurred during event handling: %v", err)
	}

	// Check if the transition was cached after the first call
	if len(fsm.transitionCache) == 0 {
		t.Errorf("Transition was not cached as expected")
	}
}

func TestFSM_HandleEventImpl_Callbacks(t *testing.T) {
	// Create a new FSM for testing
	fsm := NewFSM("zero", nil)

	// Set up callback flags
	onLeaveCalled := false
	beforeEventCalled := false
	afterEventCalled := false

	// Set BeforeEvent and AfterEvent callbacks
	fsm.BeforeEvent = func(ev *Event) {
		beforeEventCalled = true
	}
	fsm.AfterEvent = func(ev *Event, err error) {
		afterEventCalled = true
	}
	// Set onLeave callback
	onLeaveCallback := func(ev *Event) error {
		onLeaveCalled = true
		return fmt.Errorf("error 16")
	}

	// Define an eventInfo for testing
	eventPlusOne := "plus-one"

	// Add the transitions to the FSM
	fsm.AddTransition(NewTransition("zero", "one", eventPlusOne, nil, nil, onLeaveCallback))
	fsm.AddTransition(NewTransition("one", "two", eventPlusOne, nil, nil, nil))

	// Call the public handleEvent function
	err := fsm.HandleEvent(eventPlusOne, nil)
	assert.NoError(t, err, "Error occurred during event handling")
	// Check last event
	assert.Equal(t, eventPlusOne, fsm.LastEvent())

	err = fsm.HandleEvent(eventPlusOne, nil)
	// Check if onLeave callback was called with error
	assert.Error(t, err, "error 16")
	// Check if onLeave callback was called
	assert.True(t, onLeaveCalled, "onLeave callback was not called as expected")

	// Check if BeforeEvent callback was called
	assert.True(t, beforeEventCalled, "BeforeEvent callback was not called as expected")

	// Check if AfterEvent callback was called
	assert.True(t, afterEventCalled, "AfterEvent callback was not called as expected")
}

func TestHandleEventImplWithAnyTransition(t *testing.T) {
	// Create a mock FSM
	fsm := &FSM{
		currentState:    "state1",
		currentEvent:    "event1",
		transitionCache: make(map[transitionKey]*Transition),
	}

	onEvent := func(e *Event) error {
		return nil
	}
	NewTransitionBuilder().
		From(SuperStateAny).
		To("state2").
		Event("event1").
		OnEvent(onEvent).
		BuildAndAdd(fsm)

	// Create an event to test with
	eventInfo := eventInfo{
		event: "event1",
		data:  nil,
	}

	// Call the handleEventImpl() function
	err := fsm.handleEventImpl(eventInfo)
	// Check if the transition was successful
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	// Check if the current state has been updated
	if fsm.currentState != "state2" {
		t.Errorf("Expected currentState to be 'state2', got %v", fsm.currentState)
	}
}

func TestHandleEventImplWithAnyTransitionNoTransition(t *testing.T) {
	// Create a mock FSM with no matching transition
	fsm := &FSM{
		currentState:    "state1",
		currentEvent:    "event1",
		transitionCache: make(map[transitionKey]*Transition),
	}

	// Create an event to test with
	eventInfo := eventInfo{
		event: "event1",
		data:  nil,
	}

	// Call the handleEventImpl() function
	err := fsm.handleEventImpl(eventInfo)

	// Check if the expected error (ErrEventNotSupported) is returned
	if err != ErrEventNotSupported {
		t.Errorf("Expected ErrEventNotSupported error, got %v", err)
	}

	// Check if the current state has not changed
	if fsm.currentState != "state1" {
		t.Errorf("Expected currentState to be 'state1', got %v", fsm.currentState)
	}
}
