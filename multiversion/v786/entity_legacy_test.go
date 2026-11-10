package v786

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// The flag numbers of Dragonfly's own 1.16/1.19-era sessions: breathing 35,
// has collision 47, affected by gravity 48, swimming 56.
func TestLegacyEntityFlagsAndKeys(t *testing.T) {
	const breathing, collision, gravity, swimming, newFlag = 35, 48, 49, 57, 46
	in := protocol.EntityMetadata{
		protocol.EntityDataKeyFlags: int64(1<<breathing | 1<<collision | 1<<gravity | 1<<swimming | 1<<newFlag),
		125:                         int32(7), // a key 527 has one slot further
		135:                         int32(1), // a key 527 does not have
		4:                           "nametag",
	}
	for _, proto := range []uint32{527, 407} {
		out := legacyEntityMetadata(proto, in)
		want := int64(1<<35 | 1<<47 | 1<<48 | 1<<56)
		if got := out[protocol.EntityDataKeyFlags].(int64); got != want {
			t.Errorf("%d: flags %b, want %b", proto, got, want)
		}
		if out[4] != "nametag" {
			t.Errorf("%d: name tag key moved", proto)
		}
		if _, ok := out[135]; ok {
			t.Errorf("%d: kept a key it has no slot for", proto)
		}
	}
	if out := legacyEntityMetadata(527, in); out[126] != int32(7) {
		t.Errorf("527: key 125 did not move to 126: %v", out)
	}
	if out := legacyEntityMetadata(589, in); out[protocol.EntityDataKeyFlags] != in[protocol.EntityDataKeyFlags] {
		t.Error("589 is not a legacy protocol and must be untouched")
	}
}
