package legacy

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestEveryProtocolHasKnownIDs(t *testing.T) {
	for _, id := range order {
		if len(knownIDs[id]) < 100 {
			t.Errorf("protocol %d: %d known packet IDs", id, len(knownIDs[id]))
		}
	}
}

func TestUnknownPacketsAreDropped(t *testing.T) {
	p := New(582)
	if out := p.ConvertFromLatest(&packet.LocatorBar{}, nil); len(out) != 0 {
		t.Fatalf("1.19.80 got %T, a packet from 1.21.80", out[0])
	}
	if out := p.ConvertFromLatest(&packet.SetTime{Time: 1}, nil); len(out) != 1 {
		t.Fatal("SetTime dropped")
	}
}
