package peers

import (
	"testing"
	"time"
)

func TestBackoffer(t *testing.T) {
	tests := []struct {
		name         string
		maxAttempts  int
		attemptDelay time.Duration
		baseDelay    time.Duration
		sequence     []bool
	}{
		{
			name:         "first attempt always succeeds",
			maxAttempts:  3,
			attemptDelay: 20 * time.Millisecond,
			baseDelay:    20 * time.Millisecond,
			sequence:     []bool{true},
		},
		{
			name:         "max attempts reached",
			maxAttempts:  3,
			attemptDelay: 2 * time.Millisecond,
			baseDelay:    1 * time.Millisecond,
			sequence:     []bool{true, true, true, false, false},
		},
		{
			name:         "respects backoff delay",
			maxAttempts:  5,
			attemptDelay: 30 * time.Millisecond,
			baseDelay:    50 * time.Millisecond,
			sequence:     []bool{true, false, true, false, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &backoffer{
				maxAttempts: tt.maxAttempts,
				baseDelay:   tt.baseDelay,
				maxDelay:    time.Minute,
				// no jitter in these tests
			}

			for i, expected := range tt.sequence {
				if i > 0 {
					// Sleep a variable amount based on the attempt number
					time.Sleep(tt.attemptDelay * time.Duration(i))
				}

				result := b.try()
				if result != expected {
					t.Errorf("attempt %d: got %v, want %v", i+1, result, expected)
				}
			}
		})
	}
}

func TestBackofferJitter(t *testing.T) {
	// first with no jitter
	b := newBackoffer(8, 50*time.Millisecond, time.Minute, false)

	t0 := time.Now()
	for range 3 {
		for !b.try() {
			time.Sleep(5 * time.Millisecond)
		}
	}
	totalDelayNoJitter := time.Since(t0)

	// again with jitter enabled, should be longer in total
	b = newBackoffer(8, 50*time.Millisecond, time.Minute, true)
	t0 = time.Now()
	for range 3 {
		for !b.try() {
			time.Sleep(5 * time.Millisecond)
		}
	}
	totalDelayJitter := time.Since(t0)

	// total delay with jitter should be longer than without
	if totalDelayJitter < totalDelayNoJitter {
		t.Errorf("total delay with jitter (%v) should be longer than without jitter (%v)", totalDelayJitter, totalDelayNoJitter)
	}
	t.Log(totalDelayJitter - totalDelayNoJitter)
}

func TestCalculateBackoffTTL(t *testing.T) {
	tests := []struct {
		name      string
		base      time.Duration
		max       time.Duration
		retries   int
		jitter    bool
		expected  time.Duration
		checkFunc func(t *testing.T, result time.Duration)
	}{
		{
			name:     "zero retries",
			base:     time.Second,
			max:      time.Hour,
			retries:  0,
			jitter:   false,
			expected: 0,
		},
		{
			name:     "single retry no jitter",
			base:     time.Second,
			max:      time.Hour,
			retries:  1,
			jitter:   false,
			expected: time.Second,
		},
		{
			name:     "multiple retries with max cap",
			base:     time.Minute,
			max:      2 * time.Minute,
			retries:  3,
			jitter:   false,
			expected: 5 * time.Minute,
			checkFunc: func(t *testing.T, result time.Duration) {
				if result > 6*time.Minute {
					t.Errorf("backoff exceeded expected maximum: got %v", result)
				}
			},
		},
		{
			name:    "jitter impact verification",
			base:    time.Second,
			max:     time.Hour,
			retries: 4,
			jitter:  true,
			checkFunc: func(t *testing.T, result time.Duration) {
				noJitterResult := calculateBackoffTTL(time.Second, time.Hour, 4, false)
				if result <= noJitterResult {
					t.Errorf("expected jitter to increase total backoff time: with=%v, without=%v", result, noJitterResult)
				}
			},
		},
		{
			name:    "max delay boundary",
			base:    time.Minute,
			max:     5 * time.Minute,
			retries: 5,
			jitter:  false,
			checkFunc: func(t *testing.T, result time.Duration) {
				if result > 25*time.Minute {
					t.Errorf("backoff exceeded maximum delay boundary: got %v", result)
				}
			},
		},
		{
			name:    "large retry count",
			base:    time.Second,
			max:     time.Hour,
			retries: 63,
			jitter:  false,
			checkFunc: func(t *testing.T, result time.Duration) {
				if result < time.Hour*10 {
					t.Errorf("expected large backoff for high retry count: got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateBackoffTTL(tt.base, tt.max, tt.retries, tt.jitter)

			if tt.expected > 0 && result != tt.expected {
				t.Errorf("calculateBackoffTTL() = %v, want %v", result, tt.expected)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}
