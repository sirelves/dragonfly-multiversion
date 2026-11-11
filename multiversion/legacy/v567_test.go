package legacy

import (
	"bytes"
	"testing"

	"github.com/df-mc/dragonfly/multiversion/v786"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

func TestBefore11962(t *testing.T) {
	for v, want := range map[string]bool{"1.19.60": true, "1.19.61": true, "1.19.62": false, "1.19.63": false, "1.20.0": false, "": false} {
		if got := before11962(v); got != want {
			t.Errorf("before11962(%q) = %v, want %v", v, got, want)
		}
	}
}

func TestSkinOverrideFlagByte(t *testing.T) {
	enc := func(s v786.Skin786) int {
		buf := new(bytes.Buffer)
		w := New(567).NewWriter(buf, 0)
		protocol.Single(w, &s)
		return buf.Len()
	}
	if with, without := enc(v786.Skin786{}), enc(v786.Skin786{NoOverrideAppearance: true}); with != without+1 {
		t.Fatalf("1.19.62 skin is %d bytes, 1.19.60 skin %d: want exactly the flag apart", with, without)
	}
}
