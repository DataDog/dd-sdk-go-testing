package utils

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
)

var parentContextKey = "parentContextKey"

func WithParent(ctx context.Context, parent context.Context) context.Context {
	return context.WithValue(ctx, &parentContextKey, parent)
}

func GetParent(ctx context.Context) context.Context {
	parentUntyped := ctx.Value(&parentContextKey)
	if parentUntyped == nil {
		return nil
	}

	parent, ok := parentUntyped.(context.Context)
	if !ok {
		panic(fmt.Errorf("GetParent: parent is not a context in %+v", ctx))
	}

	return parent
}

var finishContextKey = "finishContextKey"

func WithFinish(ctx context.Context, finish func()) context.Context {
	return context.WithValue(ctx, &finishContextKey, finish)
}

func Finish(ctx context.Context) error {
	finishUntyped := ctx.Value(&finishContextKey)
	if finishUntyped == nil {
		return errors.New("Finish: no finish function")
	}

	finish, ok := finishUntyped.(func())
	if !ok {
		return errors.New("Finish: finish is not a function")
	}

	finish()

	return nil
}

var testCountContextKey = "testCountContextKey"

func WithTestCount(ctx context.Context, count int64) context.Context {
	return context.WithValue(ctx, &testCountContextKey, &count)
}

func IncTestCount(ctx context.Context) int64 {
	countUntyped := ctx.Value(&testCountContextKey)
	if countUntyped == nil {
		panic(fmt.Errorf("IncTestCount: no test count in %+v", ctx))
	}

	count, ok := countUntyped.(*int64)
	if !ok {
		panic(fmt.Errorf("IncTestCount: test count is not a pointer in %+v", ctx))
	}

	return atomic.AddInt64(count, 1)
}

func DecTestCount(ctx context.Context) int64 {
	countUntyped := ctx.Value(&testCountContextKey)
	if countUntyped == nil {
		panic(fmt.Errorf("DecTestCount: no test count in %+v", ctx))
	}

	count, ok := countUntyped.(*int64)
	if !ok {
		panic(fmt.Errorf("DecTestCount: test count is not a pointer in %+v", ctx))
	}

	return atomic.AddInt64(count, -1)
}

func GetTestCount(ctx context.Context) int64 {
	countUntyped := ctx.Value(&testCountContextKey)
	if countUntyped == nil {
		panic(fmt.Errorf("GetTestCount: no test count in %+v", ctx))
	}

	count, ok := countUntyped.(*int64)
	if !ok {
		panic(fmt.Errorf("GetTestCount: test count is not a pointer in %+v", ctx))
	}

	return *count
}
