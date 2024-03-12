package autoinstrument

import (
	"context"
	ddtesting "github.com/DataDog/dd-sdk-go-testing"
	"reflect"
	"sync"
	"testing"
	"unsafe"
)

var (
	contextMutex sync.RWMutex
	contextMap   = map[*testing.T]context.Context{}
)

// Implementation for auto instrumentation

func Run(t *testing.T, name string, f func(t *testing.T)) bool {
	return t.Run(name, func(t *testing.T) {
		_, finish := ddtesting.StartTestWithContext(GetContext(t), t, ddtesting.WithOriginalTestFunc(f))
		defer finish()
		f(t)
	})
}

func RunM(m *testing.M) int {

	// Let's access to the inner Test array and instrument them
	internalTests := getInternalTestArray(m)
	if internalTests != nil {
		newTestArray := make([]testing.InternalTest, len(*internalTests))
		for idx, test := range *internalTests {
			testFn := test.F
			newTestArray[idx] = testing.InternalTest{
				Name: test.Name,
				F: func(t *testing.T) {
					_, finish := ddtesting.StartTestWithContext(GetContext(t), t, ddtesting.WithOriginalTestFunc(testFn))
					defer finish()
					testFn(t)
				},
			}
		}
		*internalTests = newTestArray
	}

	return ddtesting.Run(m)
}

// get the pointer to the internal test array
func getInternalTestArray(m *testing.M) *[]testing.InternalTest {
	indirectValue := reflect.Indirect(reflect.ValueOf(m))
	member := indirectValue.FieldByName("tests")
	if member.IsValid() {
		return (*[]testing.InternalTest)(unsafe.Pointer(member.UnsafeAddr()))
	}
	return nil
}

func GetContext(t *testing.T) context.Context {
	// Read lock
	contextMutex.RLock()
	if ctx, ok := contextMap[t]; ok {
		return ctx
	}
	contextMutex.RUnlock()

	// Write lock
	ctx := context.Background()
	contextMutex.Lock()
	contextMap[t] = ctx
	contextMutex.Unlock()
	return ctx
}
