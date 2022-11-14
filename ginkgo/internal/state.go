package utils

import (
	"context"
	"errors"
)

type State struct {
	ctx context.Context
}

func NewState(ctx context.Context) *State {
	return &State{
		ctx: ctx,
	}
}

func (s *State) Push(ctx context.Context) {
	s.ctx = WithParent(ctx, s.ctx)
}

func (s *State) Pop() error {
	parent := GetParent(s.ctx)
	if parent == nil {
		return errors.New("no parent")
	}

	s.ctx = parent

	return nil
}

// GetContexts returns all contexts in the current stack.
func (s *State) GetContexts() []context.Context {
	result := []context.Context{}

	for ctx := s.ctx; ctx != nil; ctx = GetParent(ctx) {
		result = append(result, ctx)
	}

	return result
}
