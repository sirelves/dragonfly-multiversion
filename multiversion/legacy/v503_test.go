package legacy

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestOldFlightReports(t *testing.T) {
	st := &connState{}
	up := delta503.client(503, &packet.AdventureSettings{Flags: packet.AdventureFlagFlying | packet.AdventureFlagAllowFlight}, st)
	if r, ok := up[0].(*packet.RequestAbility); !ok || r.Ability != packet.AbilityFlying {
		t.Fatalf("taking off became %#v", up)
	}
	if out := delta503.client(503, &packet.AdventureSettings{Flags: packet.AdventureFlagAllowFlight}, st); len(out) != 0 {
		t.Fatalf("landing produced %d packets", len(out))
	}
	in := &packet.PlayerAuthInput{InputData: protocol.NewInputFlags(packet.InputFlagCount)}
	delta503.latest(503, in, st)
	if !in.InputData.Load(packet.InputFlagStopFlying) {
		t.Fatal("landing did not reach the next PlayerAuthInput")
	}
	again := &packet.PlayerAuthInput{InputData: protocol.NewInputFlags(packet.InputFlagCount)}
	delta503.latest(503, again, st)
	if again.InputData.Load(packet.InputFlagStopFlying) {
		t.Fatal("StopFlying repeated on a later input")
	}
}
