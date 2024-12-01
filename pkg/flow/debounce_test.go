package flow_test

import (
	"testing"
	"time"

	"github.com/hans-m-song/watchdog/pkg/flow"
	"github.com/stretchr/testify/assert"
)

func TestDebounce(t *testing.T) {
	calls := int32(0)
	call := func() { calls += 1 }
	debounce := flow.NewDebouncer(20 * time.Millisecond)

	debounce(call)
	debounce(call)
	debounce(call)
	time.Sleep(30 * time.Millisecond)
	assert.Equal(t, int32(1), calls)

	debounce(call)
	debounce(call)
	debounce(call)
	time.Sleep(30 * time.Millisecond)
	assert.Equal(t, int32(2), calls)

	debounce(call)
	debounce(call)
	debounce(call)
	time.Sleep(30 * time.Millisecond)
	assert.Equal(t, int32(3), calls)
}
