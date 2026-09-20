package quic

import (
	"bytes"
	"testing"
)

func TestAssembleCryptoFrames(t *testing.T) {
	frames := []cryptoFrame{
		{Offset: 2, Data: []byte("cd")},
		{Offset: 0, Data: []byte("ab")},
	}
	if got := assembleCryptoFrames(frames); !bytes.Equal(got, []byte("abcd")) {
		t.Fatalf("assembled frames = %q, want abcd", got)
	}

	gap := []cryptoFrame{
		{Offset: 0, Data: []byte("ab")},
		{Offset: 3, Data: []byte("d")},
	}
	if got := assembleCryptoFrames(gap); got != nil {
		t.Fatalf("assembled noncontiguous frames = %q, want nil", got)
	}
}
