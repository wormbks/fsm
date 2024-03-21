package callchain

import (
	"context"
)

// State is a type representing the state of the finite state machine.
type State[T any] func(ctx context.Context, args Args[T]) (Args[T], error)

// Args is a generic struct that contains data and a next state function for use in a state machine.
type Args[T any] struct {
	Data T
	Next State[T]
}


// Run executes the state machine starting from the provided start State.
// It runs each state until completion, error, or context cancellation.
// The final state's Args are returned, containing the final data T.
func Run[T any](ctx context.Context, arg T, start State[T]) (T, error) {
	var err error
	args := Args[T]{Data: arg, Next: start}
	current := start
	args.Next = nil // Clear this so to prevent infinite loops
	for {
		if ctx.Err() != nil {
			return args.Data, ctx.Err()
		}
		args, err = current(ctx, args)
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

// Fsm type to manage FSM instances
type Fsm[T any] struct {
  Start State[T]
  Data  T
}

// Method to run the FSM with a given context
func (f *Fsm[T]) Run(ctx context.Context) (T, error) {
  return Run(ctx, f.Data, f.Start)
}

