package tiny

import "context"

// Event is a type that can represent any value. It is used to
// pass events between state machines and channels.
type Event any

// Args is a struct that contains arguments passed to state functions.
// The generic type parameter T allows the arguments to be of any type.
type Args[T any] struct {
	Data T
	Next State[T]
  }
  
// State is a function that handles a state transition in a finite state machine.
// It takes a context, the current state arguments, a channel for receiving events,
// and returns the next state arguments or an error.
type State[T any] func(ctx context.Context, args Args[T], eventChannel <-chan Event) (Args[T], error)

// Run executes the provided state machine start State in the given context,
// passing the initial arg T. It returns the final state's arg and any error.
// Run handles looping through each state transition until a nil next State is
// returned or the context is canceled.
func Run[T any](ctx context.Context, arg T, start State[T], eventChannel <-chan Event) (T, error) {
	var err error

	args := Args[T]{Data: arg, Next: start}
	current := start
	args.Next = nil // Clear this so to prevent infinite loops
	for {
		if ctx.Err() != nil {
			return args.Data, ctx.Err()
		}
		args, err = current(ctx, args, eventChannel)
		if err != nil {
			return args.Data, err
		}
		current = args.Next // Set our next stage
		args.Next = nil     // Clear this so to prevent infinite loops

		if current == nil {
			return args.Data, nil
		}
	}
}
  
// FsmWithEvents is a state machine that receives events
// from an external channel. It allows the state machine logic to be
// separated from the event source.
type FsmWithEvents[T any] struct {
	Start State[T]
	Data  T
	EventChannel chan Event
  }
  
  // Method to run the FSM with a given context
  func (f *FsmWithEvents[T]) Run(ctx context.Context) (T, error) {
	return Run(ctx, f.Data, f.Start, f.EventChannel)
  }
  