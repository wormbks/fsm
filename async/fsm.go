package async

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	defChannelBufferSize = 1024 // it should be big enough to avoid blocking
	SuperStateAny        = "ANY_STATE"
)

var (
	// This error is returned when a transition is not allowed
	ErrTransitionNotAllowed = fmt.Errorf("transition not allowed")
	// This error is returned when a transition should not be performed
	// because the data received by Action. For example, a transition to
	// a different state should be performed if the data received contains
	// a different value or some flag/signal.
	ErrNoTransition = fmt.Errorf("no transition")
	// This error is returned when an event is not supported for the current state
	ErrEventNotSupported = fmt.Errorf("event not supported in the current state")
	// This error is returned when a timeout occurs
	ErrTimeout = fmt.Errorf("timeout")

	ErrFsmError = fmt.Errorf("fsm error")
)

type ErrorReporter interface {
	SendError(err error)
}

type Event struct {
	Event string
	From  string
	To    string
	Data  interface{}
}

// FSM represents a Finite State Machine.
type FSM struct {
	initialState string // Represents the initial state of the FSM.
	currentState string // Represents the current state of the FSM.
	currentEvent string // Represents the current event being processed by the FSM.

	eventHandleMutex  sync.Mutex // A mutex used to synchronize access to event handling code.
	eventWritingMutex sync.Mutex // A mutex used to synchronize writing to the event channel.

	eventChannel    chan eventInfo                // A channel used to send event information to the FSM.
	errorReporter   ErrorReporter                 // A callback used to send errors from the FSM.
	lastTransition  *Transition                   // Stores the last transition that occurred in the FSM.
	once            sync.Once                     // A sync.Once object used for stoping the FSM.
	transitionCache map[transitionKey]*Transition // A map used to cache transitions for performance optimization.

	// It is used for debug callbacks mostly.
	BeforeEvent func(ev *Event)            // A callback function that is called before an event is processed. If nil, no action is performed.
	AfterEvent  func(ev *Event, err error) // A callback function that is called after an event is processed. If nil, no action is performed.
	// Watchdog callback
	OnWatchdog func(fsm *FSM) bool // A callback function that is called periodically to check if the FSM should be stopped. If nil, no action is performed. If true, the FSM is stopped.
}

// NewFSM creates a new FSM instance with the initial state.
func NewFSM(initialState string, errCb ErrorReporter) *FSM {
	return &FSM{
		initialState: initialState,
		currentState: initialState,
		// transitions:     make([]*Transition, 0, defTransitionBuffer),
		eventHandleMutex: sync.Mutex{},
		eventChannel:     make(chan eventInfo, defChannelBufferSize),
		errorReporter:    errCb,
		lastTransition:   nil,
		transitionCache:  make(map[transitionKey]*Transition),
	}
}

// AddTransition adds a transition to the FSM.
func (fsm *FSM) AddTransition(transition *Transition) *FSM {
	fsm.eventHandleMutex.Lock()
	defer fsm.eventHandleMutex.Unlock()
	fsm.transitionCache[makeTransitionKey(transition.Event, transition.From)] = transition
	return fsm
}

// Start starts the FSM in a new goroutine.
//
// It takes a WaitGroup as a parameter to coordinate the completion of the goroutine.
// It does not return any value.
func (fsm *FSM) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		fsm.handleEvents()
	}()
}

// Stop stops the FSM.
func (fsm *FSM) Stop() {
	fsm.once.Do(func() {
		fsm.eventWritingMutex.Lock()
		defer fsm.eventWritingMutex.Unlock()
		if fsm.eventChannel != nil {
			close(fsm.eventChannel)
			fsm.eventChannel = nil
		}
	})
}

// Run runs the FSM (Finite State Machine) with a given context, timeout, and ignoreNotSupported flag.
//
// ctx: The context to control the FSM's execution.
// timeout: The duration to wait for events before triggering the watchdog timer.
// ignoreNotSupported: A flag to indicate whether to ignore events that are not supported by the FSM.
// Returns an error if the FSM encounters any issues during execution.
func (fsm *FSM) Run(ctx context.Context,
	timeout time.Duration,
	ignoreNotSupported bool,
	enableWatchdog bool,
) (err error) {
	// Create a timer that starts immediately.
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	// log.Debug().Str("state", timeout.String()).Msg("FSM timer created")
	resetTimer := func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(timeout)
	}

	for fsm.eventChannel != nil {
		select {
		case <-ctx.Done():
			fsm.Stop()

		case eventInfo, ok := <-fsm.eventChannel:
			if !ok {
				break
			}
			err = fsm.handleEventImpl(eventInfo)
			// FSM is ok, reset the timer
			if err == nil || err == ErrNoTransition {
				resetTimer()
			}
			if ignoreNotSupported && err == ErrEventNotSupported {
				err = nil
			}
		case <-timer.C: // Watchdog timer triggered, you can take action here.
			// If fsm.OnWatchdog is defined and returns true, stop the FSM.
			if enableWatchdog && (fsm.OnWatchdog == nil || fsm.OnWatchdog(fsm)) {
				fsm.Stop()
				err = fmt.Errorf("FSM watchdog triggered")
			} else {
				// Reset the timer when an event is received.
				resetTimer()
			}

		}
		fsm.SendError(err)
	}
	return err
}

// SendError sends an error using the FSM.
//
// It takes an error parameter and does not return anything.
func (fsm *FSM) SendError(err error) {
	if err != nil && fsm.errorReporter != nil && err != ErrNoTransition {
		fsm.errorReporter.SendError(err)
	}
}

// handleEvents handles the events received on the event channel and calls the
// appropriate event handler function. It continuously listens for events on the
// channel until it is closed.
//
// This function does not take any parameters.
// It does not return any values.
func (fsm *FSM) handleEvents() {
	for fsm.eventChannel != nil {
		eventInfo, ok := <-fsm.eventChannel
		if !ok {
			fsm.errorReporter = nil
			return
		}
		err := fsm.handleEventImpl(eventInfo)
		if err != nil {
			fsm.errorReporter.SendError(err)
		}

	}
}

// HandleEvent handles the given event with the provided data.
//
// Parameters:
// - event: the name of the event to handle.
// - data: the data associated with the event.
//
// Returns:
// - an error if the event handling fails.
func (fsm *FSM) HandleEvent(event string, data interface{}) error {
	ei := eventInfo{
		event: event,
		data:  data,
	}
	return fsm.handleEventImpl(ei)
}

func (fsm *FSM) findTransition(event string) *Transition {
	// Create a key for the "any" super state
	anyKey := makeTransitionKey(event, SuperStateAny)

	// Check if transition from "any" state is cached
	anyTransition, anyFound := fsm.transitionCache[anyKey]

	if anyFound {
		return anyTransition
	}

	// Create a key to check the transition cache
	key := makeTransitionKey(event, fsm.currentState)
	// Check if transition is cached for the current state
	transition, found := fsm.transitionCache[key]

	// Use the transition from "any" state if no transition is found for the current state
	if !found {
		transition = nil
	}

	return transition
}

// callAfterEvent calls the AfterEvent function if it is not nil.
//
// It takes in an Event pointer and an error as parameters.
// There is no return type for this function.
func (fsm *FSM) callAfterEvent(ev *Event, err error) {
	if fsm.AfterEvent != nil {
		fsm.AfterEvent(ev, err)
	}
}

// handleEventImpl handles the event and updates the FSM state accordingly.
//
// It locks the FSM for thread safety and sets the current event. It checks if
// there are any super transitions for the current event. If no transition is
// found, it returns an error. Then, it creates an Event instance and calls the
// BeforeEvent function if defined. It also calls the onEvent function if
// defined and checks if a state change is going to happen. If the previous
// transition had an onLeave function, it calls that function. It then calls
// the onEnter function if defined. Finally, it updates the current state and
// last transition and calls the AfterEvent function if defined. It returns any
// error that occurred during the event handling process.
//
// Parameter(s):
// - ei: the event information containing the event and data
//
// Return type(s):
// - error: any error that occurred during the event handling process
func (fsm *FSM) handleEventImpl(ei eventInfo) error {
	// Lock the FSM for thread safety
	fsm.eventHandleMutex.Lock()
	defer fsm.eventHandleMutex.Unlock()
	// Set the current event
	fsm.currentEvent = ei.event
	// Do we have any super transitions for the current event?
	transition := fsm.findTransition(ei.event)
	// If no transition found, return an error
	if transition == nil {
		return ErrEventNotSupported
	}
	var err error
	// Create the Event instance
	ev := Event{
		Event: ei.event,
		Data:  ei.data,
		From:  transition.From,
		To:    transition.To,
	}
	// Call BeforeEvent if defined
	if fsm.BeforeEvent != nil {
		fsm.BeforeEvent(&ev)
	}
	// Call onEvent if defined
	if transition.onEvent != nil {
		err = transition.onEvent(&ev)
		if err != nil {
			if err != ErrNoTransition {
				err = fmt.Errorf("on event %s from %s : %w", ev.Event, ev.From, err)
			}
			fsm.callAfterEvent(&ev, err)
			return err
		}
	}

	// Check if a state change is going to happen
	if fsm.currentState != transition.To {
		// Check if the previous transition had an onLeave function
		if fsm.lastTransition != nil && fsm.lastTransition.onLeave != nil {
			err = fsm.lastTransition.onLeave(&ev)
			if err != nil {
				fsm.callAfterEvent(&ev, err)
				return fmt.Errorf("on leave error: %w", err)
			}
		}
	}

	// Call onEnter if defined
	if transition.onEnter != nil {
		err = transition.onEnter(&ev)
		if err != nil {
			fsm.callAfterEvent(&ev, err)
			return fmt.Errorf("on enter error: %w", err)
		}
	}

	// Update the current state and last transition
	fsm.currentState = transition.To
	fsm.lastTransition = transition
	// call to AfterEvent if defined
	fsm.callAfterEvent(&ev, err)

	// Return any error that occurred during the event handling process
	return err
}

// handleEventImpl triggers an event on the FSM with optional parameters.

// CurrentState returns the current state of the FSM.
func (fsm *FSM) CurrentState() string {
	return fsm.currentState
}

// PreviousState returns the previous state of the FSM.
func (fsm *FSM) PreviousState() string {
	if fsm.lastTransition != nil {
		return fsm.lastTransition.To
	}

	return fsm.initialState
}

// LastEvent returns the current event triggered in the FSM.
func (fsm *FSM) LastEvent() string {
	return fsm.currentEvent
}

// SendEvent sends an event to the FSM (Finite State Machine).
//
// It takes in the event string and params interface{} as parameters.
// It returns an error, if any was in the error channel. In case of an error
// is returned and the event was not sent.
func (fsm *FSM) SendEvent(event string, data interface{}) error {
	evInf := eventInfo{event: event, data: data}
	fsm.eventWritingMutex.Lock()
	defer fsm.eventWritingMutex.Unlock()

	select {
	case fsm.eventChannel <- evInf:
	default:
		return fmt.Errorf("can not send event state %s", fsm.currentState)

	}
	return nil
}

// GenerateMermaidDiagram generates a Mermaid diagram representing the FSM.
func GenerateMermaidDiagram(fsm *FSM, showFunc bool) string {
	var builder strings.Builder

	// Start the Mermaid diagram definition
	builder.WriteString("``` mermaid\n")
	builder.WriteString("stateDiagram\n")

	// Add the initial state
	builder.WriteString(fmt.Sprintf("[*] --> %s\n", fsm.initialState))

	// Add the states and transitions
	for _, t := range fsm.transitionCache {
		builder.WriteString(fmt.Sprintf("%s --> %s : %s\n", t.From, t.To, t.Event))
	}

	builder.WriteString("END --> [*]\n")

	// End the Mermaid diagram definition
	builder.WriteString("```\n")

	return builder.String()
}

type eventInfo struct {
	event string
	data  interface{}
}
