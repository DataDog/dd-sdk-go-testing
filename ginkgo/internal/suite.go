package utils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	ddtesting "github.com/DataDog/dd-sdk-go-testing"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/dsl/reporting"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type SuiteTest struct {
	sync.Mutex

	TestFrameworkName string
	state             State
	closed            bool
}

func NewSuiteTest(frameworkName string) *SuiteTest {
	return &SuiteTest{
		TestFrameworkName: frameworkName,
		state:             *NewState(WithTestCount(WithFinish(context.Background(), func() {}), 0)),
		closed:            false,
	}
}

func (s *SuiteTest) Context() context.Context {
	return s.state.ctx
}

func (s *SuiteTest) EnterContainer(stepName string, text string) error {
	s.Lock()
	defer s.Unlock()

	if s.closed {
		return errors.New("suite is closed")
	}

	span, ctx := tracer.StartSpanFromContext(s.state.ctx, text,
		tracer.Tag("test.name", text),
		tracer.Tag("test.type", "test"),
		tracer.Tag("test.framework", s.TestFrameworkName),
		tracer.Tag("ginkgo.seed", ginkgo.GinkgoRandomSeed()),
		tracer.Tag("ginkgo.step", stepName),
	)

	s.state.Push(WithTestCount(WithFinish(ctx, func() { span.Finish() }), 0))

	return nil
}

func (s *SuiteTest) LeaveContainer() error {
	s.Lock()
	defer s.Unlock()

	if s.closed {
		return errors.New("suite is closed")
	}

	if err := s.state.Pop(); err != nil {
		return fmt.Errorf("no parent for %+v: %w", s.state.ctx, err)
	}

	return nil
}

type test struct {
	ctx     context.Context
	finish  func()
	options []ddtesting.Option
}

func newTest(ctx context.Context, opts ...ddtesting.Option) *test {
	for ctx := ctx; ctx != nil; ctx = GetParent(ctx) {
		if IncTestCount(ctx) < 1 {
			panic(fmt.Errorf("corrupted stack on %+v", ctx))
		}
	}

	return &test{ctx: ctx, options: opts}
}

func (t *test) Enter(stepName, suiteName string) error {
	opts := append(t.options, ddtesting.WithSpanOptions(
		tracer.Tag("test.type", "test"),
		tracer.Tag("test.suite", suiteName),
		tracer.Tag("ginkgo.step", stepName),
		tracer.Tag("ginkgo.seed", ginkgo.GinkgoRandomSeed()),
		tracer.Tag("ginkgo.parallelProcess", ginkgo.GinkgoParallelProcess()),
		tracer.Tag("ginkgo.numAttempts", reporting.CurrentSpecReport().NumAttempts),
	))

	_, t.finish = ddtesting.StartTestWithContext(t.ctx, ginkgo.GinkgoT(1), opts...)

	return nil
}

func (t *test) Leave() error {
	if t.finish == nil {
		return fmt.Errorf("not entered in test %+v", t.ctx)
	}

	t.finish()

	nbToPop := 0

	for ctx := t.ctx; ctx != nil; ctx = GetParent(ctx) {
		value := DecTestCount(ctx)

		switch {
		case value == 0:
			if nbToPop < 0 {
				return fmt.Errorf("corrupted stack on %+v", ctx)
			}

			nbToPop++

			if err := Finish(ctx); err != nil {
				return fmt.Errorf("context %+v: %w", ctx, err)
			}
		case value < 0:
			return fmt.Errorf("count < 0 for %+v", ctx)
		default:
			nbToPop = -1 // mark the stack as corrupted
		}
	}

	return nil
}

type TestCase interface {
	Enter(stepName, suiteName string) error
	Leave() error
}

type Snapshot struct {
	FrameworkName string
	ctx           context.Context
}

func (s *SuiteTest) Snapshot() *Snapshot {
	return &Snapshot{
		FrameworkName: s.TestFrameworkName,
		ctx:           s.state.ctx,
	}
}

func (s *Snapshot) RegisterTest() TestCase {
	return newTest(
		s.ctx,
		ddtesting.WithSpanOptions(tracer.Tag("test.framework", s.FrameworkName)),
	)
}

func (s *SuiteTest) Close() error {
	s.Lock()
	defer s.Unlock()

	for s.state.Pop() == nil {
	}

	s.closed = true

	return nil
}

var _ io.Closer = (*SuiteTest)(nil)
