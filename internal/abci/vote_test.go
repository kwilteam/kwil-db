package abci

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractDataSegments(t *testing.T) {

	t.Run("valid data", func(t *testing.T) {
		input := []byte{0x05, 0x01, 0x02, 0x0a, 0x03, 0x04}
		want := [][]byte{{0x01, 0x02, 0x0a, 0x03, 0x04}}

		got, err := ExtractDataSegments(input)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("valid data 2", func(t *testing.T) {
		input := []byte{0x03, 0x01, 0x02, 0x0a, 0x01, 0x04}
		want := [][]byte{{0x01, 0x02, 0x0a}, {0x04}}

		got, err := ExtractDataSegments(input)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("empty input", func(t *testing.T) {
		input := []byte{}
		want := [][]byte(nil)

		got, err := ExtractDataSegments(input)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("invalid varint", func(t *testing.T) {
		input := []byte{0x0a}

		_, err := ExtractDataSegments(input)
		require.Error(t, err)
	})

	t.Run("negative length", func(t *testing.T) {
		input := []byte{0xff, 0xff, 0xff, 0xff, 0xff}

		_, err := ExtractDataSegments(input)
		require.Error(t, err)
	})

	t.Run("insufficient data", func(t *testing.T) {
		input := []byte{0x0a, 0x01}

		_, err := ExtractDataSegments(input)
		require.Error(t, err)
	})

}

func TestTestVoteExt_UnmarshalBinary(t *testing.T) {
	// round-trip a message through the binary encoding
	d := &TestVoteExt{
		Msg: "test",
	}
	b, err := d.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	var d2 TestVoteExt
	err = d2.UnmarshalBinary(b)
	if err != nil {
		t.Fatal(err)
	}
	if d.Msg != d2.Msg {
		t.Fatal("not equal")
	}
}

func TestDepositVoteExt_MarshalUnmarshal(t *testing.T) {
	// round-trip a message through the binary encoding
	d := &DepositVoteExt{
		EventID: "test event ID",
		Account: "test account",
		Amount:  "10",
	}
	b, err := d.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	var d2 DepositVoteExt
	err = d2.UnmarshalBinary(b)
	if err != nil {
		t.Fatal(err)
	}
	if d.EventID != d2.EventID {
		t.Fatal("not equal")
	}
	if d.Account != d2.Account {
		t.Fatal("not equal")
	}
	if d.Amount != d2.Amount {
		t.Fatal("not equal")
	}
}

func TestEncodeDecodeVoteExtension(t *testing.T) {
	// round-trip a vote extension with two segments of different types
	var vExt []*VoteExtensionSegment

	// Add a "debug" vote extension segment
	testExt := &TestVoteExt{
		Msg: "test message",
	}
	testExtData, _ := testExt.MarshalBinary()

	vExt = append(vExt, &VoteExtensionSegment{
		Version: 0,
		Type:    VoteExtensionTypeTest,
		Data:    testExtData,
	})

	// Add a "deposit" vote extension segment
	depExt := &DepositVoteExt{
		EventID: "test event ID",
		Account: "test account",
		Amount:  "10",
	}
	depExtData, _ := depExt.MarshalBinary()

	vExt = append(vExt, &VoteExtensionSegment{
		Version: 0,
		Type:    VoteExtensionTypeDeposit,
		Data:    depExtData,
	})

	ve := EncodeVoteExtension(vExt)

	// Decode the vote extension
	vExt2, err := DecodeVoteExtension(ve)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, vExt, vExt2)

	// vExt2[0].Data
}
