package notification

import (
	"testing"
	"time"
)

func TestRetryDelayIsExponentialAndCapped(t *testing.T) {
	if got := retryDelay(1); got != 5*time.Second {
		t.Fatalf("first retry = %s", got)
	}
	if got := retryDelay(3); got != 20*time.Second {
		t.Fatalf("third retry = %s", got)
	}
	if got := retryDelay(50); got != 640*time.Second {
		t.Fatalf("capped retry = %s", got)
	}
}
