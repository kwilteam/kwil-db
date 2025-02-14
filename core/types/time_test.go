package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDuration_MarshalText(t *testing.T) {
	t.Run("marshal zero duration", func(t *testing.T) {
		d := Duration(0)
		data, err := d.MarshalText()
		require.NoError(t, err)
		require.Equal(t, "0s", string(data))
	})

	t.Run("marshal positive duration", func(t *testing.T) {
		d := Duration(2*time.Hour + 30*time.Minute)
		data, err := d.MarshalText()
		require.NoError(t, err)
		require.Equal(t, "2h30m0s", string(data))
	})

	t.Run("marshal negative duration", func(t *testing.T) {
		d := Duration(-1 * time.Minute)
		data, err := d.MarshalText()
		require.NoError(t, err)
		require.Equal(t, "-1m0s", string(data))
	})
}

func TestDuration_UnmarshalText(t *testing.T) {
	t.Run("unmarshal valid duration", func(t *testing.T) {
		var d Duration
		err := d.UnmarshalText([]byte("1h30m"))
		require.NoError(t, err)
		require.Equal(t, Duration(90*time.Minute), d)
	})

	t.Run("unmarshal zero duration", func(t *testing.T) {
		var d Duration
		err := d.UnmarshalText([]byte("0s"))
		require.NoError(t, err)
		require.Equal(t, Duration(0), d)
	})

	t.Run("unmarshal invalid duration", func(t *testing.T) {
		var d Duration
		err := d.UnmarshalText([]byte("invalid"))
		require.Error(t, err)
	})

	t.Run("unmarshal empty duration", func(t *testing.T) {
		var d Duration
		err := d.UnmarshalText([]byte(""))
		require.Error(t, err)
	})

	t.Run("unmarshal complex duration", func(t *testing.T) {
		var d Duration
		err := d.UnmarshalText([]byte("2h45m30.5s"))
		require.NoError(t, err)
		expected := Duration(2*time.Hour + 45*time.Minute + 30*time.Second + 500*time.Millisecond)
		require.Equal(t, expected, d)
	})
}

func TestDurationRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		duration Duration
	}{
		{"1 hour", Duration(time.Hour)},
		{"2 minutes", Duration(2 * time.Minute)},
		{"500 ms", Duration(500 * time.Millisecond)},
		{"90 minutes", Duration(90 * time.Minute)},
		{"zero", Duration(0)},
		{"mixed", Duration(time.Hour + 2*time.Minute + 6*time.Second)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, err := tt.duration.MarshalText()
			require.NoError(t, err)

			var decoded Duration
			err = decoded.UnmarshalText(text)
			require.NoError(t, err)

			assert.Equal(t, tt.duration, decoded)
		})
	}
}
