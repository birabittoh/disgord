package music

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"
	"testing"
)

func TestIsExpectedStreamStop(t *testing.T) {
	// Mirror how miri wraps the underlying write error through its call chain.
	wrappedEPIPE := fmt.Errorf("failed to get song content: %w",
		fmt.Errorf("failed to stream to target: %w",
			&os.PathError{Op: "write", Path: "|1", Err: syscall.EPIPE}))

	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"wrapped broken pipe", wrappedEPIPE, true},
		{"closed pipe", io.ErrClosedPipe, true},
		{"context canceled", context.Canceled, true},
		{"genuine failure", errors.New("failed to get song content: auth error"), false},
		{"deadline exceeded is not an expected stop", context.DeadlineExceeded, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isExpectedStreamStop(tc.err); got != tc.want {
				t.Errorf("isExpectedStreamStop(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
