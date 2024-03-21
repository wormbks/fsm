package tiny

import (
	"context"
	"testing"
)

func TestFsmChanRun(t *testing.T) {
	startState := func(ctx context.Context, args Args[int], eventChannel <-chan Event) (Args[int], error) {
		args.Data++
		return args, nil
	}

	fsm := FsmWithEvents[int]{
		Start: startState,
		Data: 0,	}

	expected := 1

	result, err := fsm.Run(context.Background())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result != expected {
		t.Errorf("Expected %d, got %d", expected, result)
	}
}

func TestFsmChanRunWithError(t *testing.T) {
	startState := func(ctx context.Context, args Args[int], eventChannel <-chan Event) (Args[int], error) {
		return args, context.Canceled
	}

	fsm := FsmWithEvents[int]{
		Start: startState,
	}

	_, err := fsm.Run(context.Background())
	if err == nil {
		t.Error("Expected error but got nil")
	}
}


// MockState is a mock implementation of StateChan for testing purposes
func MockState(ctx context.Context, args Args[int], eventChannel <-chan Event) (Args[int], error) {
	// Mock implementation returns incremented value of Data
	args.Data++
	return args, nil
}

func TestRunChan(t *testing.T) {
	ctx := context.Background()

	// Initialize ArgsChan with initial value
	args := Args[int]{Data: 0, Next: nil}

	// Start StateChan function to use in RunChan
	start := State[int](MockState)

	// Mock event channel
	eventChannel := make(chan Event)

	result, err := Run(ctx, args.Data, start, eventChannel)

	// Expected result is the incremented value of initial data
	expectedResult := 1

	// Check if result matches the expected result
	if result != expectedResult {
		t.Errorf("Expected result %v, got %v", expectedResult, result)
	}

	// Check if there's any error
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestFsmChan_Run(t *testing.T) {
	ctx := context.Background()

	// Initialize ArgsChan with initial value
	args := Args[int]{Data: 0, Next: nil}

	// Start StateChan function to use in Run method of FsmChan
	start := State[int](MockState)

	// Mock event channel
	eventChannel := make(chan Event)

	// Create instance of FsmChan
	fsm := FsmWithEvents[int]{Start: start, Data: args.Data, EventChannel: eventChannel}

	result, err := fsm.Run(ctx)

	// Expected result is the incremented value of initial data
	expectedResult := 1

	// Check if result matches the expected result
	if result != expectedResult {
		t.Errorf("Expected result %v, got %v", expectedResult, result)
	}

	// Check if there's any error
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// MockEvent is a mock implementation of Event for testing purposes
type MockEvent struct{}

func state1(ctx context.Context, args Args[int], eventChannel <-chan Event) (Args[int], error) {
	select {
	case <-ctx.Done():
		return args, ctx.Err()
	case <-eventChannel:
		// Increment data by 1 and transition to state2
		args.Data += 1
		args.Next = state2
		return args, nil
	}
}

func state2(ctx context.Context, args Args[int], eventChannel <-chan Event) (Args[int], error) {
	select {
	case <-ctx.Done():
		return args, ctx.Err()
	case <-eventChannel:
		// Increment data by 2 and transition to nil state
		args.Data += 2
		args.Next = nil
		return args, nil
	}
}


func TestRun(t *testing.T) {
	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a channel for events
	eventChannel := make(chan Event, 3)

	// Initialize ArgsChan with initial value
	args := Args[int]{Data: 0, Next: state1}

	// Create a mock FSM instance
	fsm := FsmWithEvents[int]{Start: state1, Data: args.Data, EventChannel: eventChannel}

	// Create a goroutine to run the FSM
	go func() {
		// Send a mock event to trigger state transitions
		eventChannel <- MockEvent{}
		eventChannel <- MockEvent{}
	}()

	// Run the FSM with the mock context
	result, err := fsm.Run(ctx)

	// Expected result after FSM run
	expectedResult := 3

	// Check if result matches the expected result
	if result != expectedResult {
		t.Errorf("Expected result %v, got %v", expectedResult, result)
	}

	// Check if there's any error
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
