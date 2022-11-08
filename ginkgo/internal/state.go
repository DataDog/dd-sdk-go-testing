package utils

import (
	"context"
	"errors"
	"fmt"

	"github.com/onsi/ginkgo/v2"
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
	fmt.Fprintln(ginkgo.GinkgoWriter, "STATE -- push")
	s.ctx = WithParent(ctx, s.ctx)
}

func (s *State) Pop() error {
	fmt.Fprintln(ginkgo.GinkgoWriter, "STATE -- pop")

	parent := GetParent(s.ctx)
	if parent == nil {
		return errors.New("state.pop: no parent")
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
