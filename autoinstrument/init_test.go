package autoinstrument

import (
	"testing"
)

func TestMain(m *testing.M) {
	RunTestMain(m)
}

func TestMyTest01(t *testing.T) {
	t.Log("My First Test")
}

func TestMyTest02(t *testing.T) {
	t.Log("My First Test 2")

	Run(t, "sub01", func(t2 *testing.T) {
		t2.Log("From sub01")

		Run(t2, "sub03", func(t3 *testing.T) {
			t3.Log("From sub03")
		})

	})
}
