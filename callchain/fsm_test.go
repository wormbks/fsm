package callchain

import (
	"context"
	"testing"
)

func TestFsmRun_func(t *testing.T) {
	nextState := func(ctx context.Context, args Args[int]) (Args[int], error) {
		args.Data++
		return args, nil
	}
	
	startState := func(ctx context.Context, args Args[int]) (Args[int], error) {
		args.Data++
		args.Next = nextState
		return args, nil
	}
	

	fsm := Fsm[int]{
		Start: startState,
			Data: 0,
		
	}
	
	expected := 2
	
	result, err := fsm.Run(context.Background())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if result != expected {
		t.Errorf("Expected %d, got %d", expected, result)
	}
}

func TestFsmRunWithError(t *testing.T) {
	startState := func(ctx context.Context, args Args[int]) (Args[int], error) {
		return args, context.Canceled
	}
	
	fsm := Fsm[int]{
		Start: startState,
	}
	
	_, err := fsm.Run(context.Background())
	if err == nil {
		t.Error("Expected error but got nil")
	}
}

type FsmSruct struct {
	Fsm[int]
	
}

func NewFsmSruct() *FsmSruct {
	res :=&FsmSruct{
		Fsm: Fsm[int]{
				Data: 0,
		},
	}
	res.Start = res.startState
	return res 
}

func (f *FsmSruct) Run(ctx context.Context) (int, error) {
	return f.Fsm.Run(ctx)
}


func (f *FsmSruct) nextState(ctx context.Context, args Args[int]) (Args[int], error) {
	args.Data+=2
	return args, nil
}


func (f *FsmSruct) startState(ctx context.Context, args Args[int]) (Args[int], error) {
	args.Data+=1
	args.Next = f.nextState
	return args, nil
}


func TestFsmRun_struct(t *testing.T) {
	

	fsm := NewFsmSruct()
	expected := 3
	
	result, err := fsm.Run(context.Background())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if result != expected {
		t.Errorf("Expected %d, got %d", expected, result)
	}
}
