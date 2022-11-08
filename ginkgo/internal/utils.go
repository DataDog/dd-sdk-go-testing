package utils

import (
	"fmt"
	"reflect"
)

func PatchFuncs(begin, end func() error, funcs ...func()) []func() {
	for i, f := range funcs {
		userFunc := f

		funcs[i] = func() {
			if err := begin(); err != nil {
				panic(fmt.Errorf("begin: %w", err))
			}

			defer func() {
				if err := end(); err != nil {
					panic(err)
				}
			}()

			userFunc()
		}
	}

	return funcs
}

func PatchArgsNoArgs(begin, end func() error, args ...interface{}) []interface{} {
	index, userFuncValue := findFunc(args...)
	if index == -1 {
		return args
	}

	patchedFunc := func() {
		if err := begin(); err != nil {
			panic(fmt.Errorf("begin: %w", err))
		}

		defer func() {
			if err := end(); err != nil {
				panic(err)
			}
		}()

		userFuncValue.Call(nil)
	}

	args[index] = patchedFunc

	return IncrementOffset(1, args...)
}

func PatchArgs(begin, end func() error, args ...interface{}) []interface{} {
	index, userFuncValue := findFunc(args...)
	if index == -1 {
		return args
	}

	patchedFunc := func(args ...interface{}) {
		if err := begin(); err != nil {
			panic(fmt.Errorf("begin: %w", err))
		}

		defer func() {
			if err := end(); err != nil {
				panic(err)
			}
		}()

		argsValues := make([]reflect.Value, len(args))
		for i, arg := range args {
			argsValues[i] = reflect.ValueOf(arg)
		}

		userFuncValue.Call(argsValues)
	}

	args[index] = patchedFunc

	return IncrementOffset(1, args...)
}

func findFunc(args ...interface{}) (int, *reflect.Value) {
	for i, arg := range args {
		if reflect.TypeOf(arg).Kind() == reflect.Func {
			v := reflect.ValueOf(arg)

			return i, &v
		}
	}

	return -1, nil
}
