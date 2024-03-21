Here's your corrected README:

# Go State Machine 

- [Go State Machine](#go-state-machine)
  - [Call Chain FSM](#call-chain-fsm)
    - [Adding External Events](#adding-external-events)
  - [Testing Advantages](#testing-advantages)
  - [Asynchronous Finite State Machine (FSM) Implementation in Go](#asynchronous-finite-state-machine-fsm-implementation-in-go)
    - [Overview](#overview)
    - [Usage](#usage)
      - [Creating an FSM](#creating-an-fsm)
      - [Adding Transitions](#adding-transitions)
      - [Starting and Stopping the FSM](#starting-and-stopping-the-fsm)
      - [Handling Events](#handling-events)
      - [Running the FSM](#running-the-fsm)
      - [Accessing FSM State Information](#accessing-fsm-state-information)
    - [Error Handling](#error-handling)
    - [Example Usage](#example-usage)
    - [Mermaid Diagram Generation](#mermaid-diagram-generation)
    - [Notes](#notes)
    - [References](#references)
    - [Transition Management](#transition-management)
      - [Transition Structure](#transition-structure)
      - [Creating Transitions](#creating-transitions)
      - [Transition Builder](#transition-builder)
        - [Methods:](#methods)
        - [Building Transitions:](#building-transitions)
      - [Transition Key Generation](#transition-key-generation)
    - [Example Usage](#example-usage-1)
  - [Notes](#notes-1)



It is based on [Go State Machine Patterns by John Doak](https://medium.com/@johnsiilver/go-state-machine-patterns-3b667f345b5e)

## Call Chain FSM

Contained within the `callchain` folder.

```go
type State[T any] func(ctx context.Context, args Args[T]) (Args[T], error)
```

This represents a function or method that receives a set of arguments of any type `T` and returns the next state and the next state function to run, or an error.

`Args` is a generic struct that contains data and a next state function for use in a state machine.

```go
type Args[T any] struct {
    Data T
    Next State[T]
}
```

If the returned `State` is `nil`, then the state machine will stop running. If an error is set, the state machine will also stop. Because you are returning what the next state to run is, you have different next states depending on various conditions.

The state machine benefits from the inclusion of generics here and the returned `T` as part of `Args`. This allows us to create a purely functional state machine (if desired) that can return a stack-allocated type that can be passed to the next state.

To make this work, we need a state runner:

```go
func Run[T any](ctx context.Context, arg T, start State[T]) (T, error) {
	var err error
	args := Args[T]{Data: arg, Next: start}
	current := start // Set our next stage
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
```

Now we have a functional state runner in a few lines of code.

### Adding External Events

Contained within the `tiny` folder.

If we want to use some external event to trigger a state change, we can do that by adding a channel event source to the state machine.

It could be done this way for a state function:

```go
type State[T any] func(ctx context.Context, args StateArgs[T], eventChannel <-chan Event) (StateArgs[T], error)
```

The `Run` function has to be modified as shown in the example:

```go
func Run[T any](ctx context.Context, args Args[T], start State[T], eventChannel <-chan Event) (T, error)
```

To handle channel events, we need to process them inside the state function:

```go
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
```

## Testing Advantages

This pattern encourages the creation of small blocks of testable code. It makes things easily divisible so that when a block gets too big, you simply create a new state that isolates the block.

But the bigger advantage is removing the need for a large end-to-end test. Because each stage in an operational flow needs to call the next stage, you end up in one of the following scenarios:

- The top-level caller calls all the sub-functions in some order.
- Each caller calls the next function.
- Some mixture of the two.

Both lead to some type of end-to-end test that shouldn’t be needed.

Here's your corrected README:

## Asynchronous Finite State Machine (FSM) Implementation in Go

The `async` package provides an implementation of a Finite State Machine (FSM) that operates asynchronously. This FSM allows for the modeling of complex behavior by defining states, transitions, and events.

### Overview

The package provides the following main components:

- **FSM**: Represents the Finite State Machine and manages its state transitions and event handling.
- **Transition**: Defines the transition between states, including the events triggering the transition and optional callback functions for handling the transition process.
- **Event**: Represents an event triggering a state transition.
- **ErrorReporter**: Interface for reporting errors encountered during FSM operation.

### Usage

#### Creating an FSM

To create a new FSM instance, use the `NewFSM` function:

```go
func NewFSM(initialState string, errCb ErrorReporter) *FSM
```

#### Adding Transitions

Use the `AddTransition` method to add transitions to the FSM:

```go
func (fsm *FSM) AddTransition(transition *Transition) *FSM
```

#### Starting and Stopping the FSM

To start the FSM, use the `Start` method, which runs the FSM in a new goroutine. Use the `Stop` method to stop the FSM:

```go
func (fsm *FSM) Start(wg *sync.WaitGroup)
func (fsm *FSM) Stop()
```

#### Handling Events

The FSM can handle events using the `HandleEvent` method:

```go
func (fsm *FSM) HandleEvent(event string, data interface{}) error
```

#### Running the FSM

The FSM can be run using the `Run` method, which controls the FSM's execution based on the provided context, timeout, and other options:

```go
func (fsm *FSM) Run(ctx context.Context, timeout time.Duration, ignoreNotSupported bool, enableWatchdog bool) error
```

#### Accessing FSM State Information

You can retrieve information about the FSM's current state, previous state, and last triggered event using the following methods:

- `CurrentState`: Returns the current state of the FSM.
- `PreviousState`: Returns the previous state of the FSM.
- `LastEvent`: Returns the last triggered event in the FSM.

### Error Handling

The FSM provides error handling mechanisms through the `ErrorReporter` interface and the `SendError` method, which reports errors encountered during FSM operation.

### Example Usage

Here's a basic example demonstrating the creation and usage of an FSM:

```go
// Create an FSM instance
fsm := async.NewFSM("initialState", nil)

// Define transitions
transition := async.NewTransition("event", "initialState", "nextState", nil, nil, nil)
fsm.AddTransition(transition)

// Start the FSM
var wg sync.WaitGroup
fsm.Start(&wg)

// Handle events
fsm.HandleEvent("event", nil)

// Stop the FSM
fsm.Stop()
```

### Mermaid Diagram Generation

The package provides a utility function `GenerateMermaidDiagram` for generating a Mermaid diagram representing the FSM's transitions.

### Notes

- This FSM implementation operates asynchronously, allowing for concurrent event handling.
- Callback functions can be provided to customize the FSM's behavior during state transitions.
- The package offers utilities for error reporting and debugging, enhancing the FSM's robustness and maintainability.

### References

- [Mermaid Documentation](https://mermaid-js.github.io/mermaid/#/)
- [Go Concurrency Patterns](https://blog.golang.org/pipelines)



Here's the README content for the Transition functionality:

### Transition Management

Transitions define the movement between states in the FSM. Each transition is associated with a specific event that triggers it, as well as optional actions to perform during the transition process.

#### Transition Structure

A transition in the FSM is represented by the `Transition` struct:

```go
type Transition struct {
    From    string          // Represents the state to transition from.
    To      string          // Represents the state to transition to.
    Event   string          // Represents the event that triggers the transition.
    Timeout time.Duration   // Represents the timeout duration for the transition.
    onEvent Action          // Represents the action to perform when the event occurs.
    onLeave Action          // Represents the action to perform when leaving the current state.
    onEnter Action          // Represents the action to perform when entering the new state.
}
```

#### Creating Transitions

To create a new transition, use the `NewTransition` function:

```go
func NewTransition(from, to, event string, onEvent, onEnter, onLeave Action) *Transition
```

#### Transition Builder

The `TransitionBuilder` struct provides a convenient way to create multiple transitions with similar properties. It allows you to specify the source state, target state, event, and actions for each transition.

##### Methods:

- `From`: Sets the source state for the transition.
- `FromMany`: Sets multiple source states for the transition.
- `To`: Sets the target state for the transition.
- `Event`: Sets the event trigger for the transition.
- `OnEvent`: Sets the action to perform when the event occurs.
- `OnEnter`: Sets the action to perform when entering the new state.
- `OnLeave`: Sets the action to perform when leaving the current state.
- `Timeout`: Sets the timeout duration for the transition.

##### Building Transitions:

- `Build`: Creates a single transition instance based on the builder's settings.
- `BuildAndAdd`: Builds the transition and adds it to the FSM instance.

#### Transition Key Generation

The `makeTransitionKey` function generates a unique key for each transition based on the event and from state. This key is used for caching transitions for performance optimization.

### Example Usage

Here's an example demonstrating the creation and usage of transitions using the Transition Builder:

```go
// Create a new TransitionBuilder
builder := async.NewTransitionBuilder()

// Define transitions
builder.From("initialState").To("nextState").Event("event").
    OnEvent(func(ev *Event) error {
        // Perform action when the event occurs
        return nil
    }).OnEnter(func(ev *Event) error {
        // Perform action when entering the new state
        return nil
    }).OnLeave(func(ev *Event) error {
        // Perform action when leaving the current state
        return nil
    })

// Add transitions to the FSM
builder.BuildAndAdd(fsm)
```

## Notes

- Transitions define the behavior of the FSM, including state changes and event handling.
- The Transition Builder simplifies the process of creating multiple transitions with consistent properties.
- Actions can be defined for each transition to customize behavior during state transitions.
- Transition keys are used for efficient caching of transitions.

---


For more information and detailed documentation, refer to the package documentation and source code.

For additional assistance, please consult the Go documentation and community resources.
