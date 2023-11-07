package gitlab

import (
	"errors"
	"fmt"
	"testing"
)

func TestTransientError(t *testing.T) {
	var stubSentinelError = errors.New("stub sentinel error")
	testCases := []struct {
		e                 error
		expectedTransient bool
		expectedFake      bool
	}{
		{
			e:                 TransientError(errors.New("test")),
			expectedTransient: true,
			expectedFake:      false,
		},
		{
			e:                 stubSentinelError,
			expectedTransient: false,
			expectedFake:      true,
		},
		{
			e:                 TransientError(stubSentinelError),
			expectedTransient: true,
			expectedFake:      true,
		},
		{
			e:                 TransientError(fmt.Errorf("%w: %w", stubSentinelError, errors.New("test"))),
			expectedTransient: true,
			expectedFake:      true,
		},
		{
			e:                 fmt.Errorf("wrapper %w", TransientError(errors.New("test"))),
			expectedTransient: true,
			expectedFake:      false,
		},
	}

	for _, tc := range testCases {
		if errors.Is(tc.e, ErrTransient) != tc.expectedTransient {
			t.Errorf("expected transient error to be %v", tc.expectedTransient)
		}
		if errors.Is(tc.e, stubSentinelError) != tc.expectedFake {
			t.Errorf("expected stub error to be %v", tc.expectedFake)
		}
	}

}
