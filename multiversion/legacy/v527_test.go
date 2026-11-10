package legacy

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func abilities(values uint32) protocol.AbilityData {
	return protocol.AbilityData{EntityUniqueID: 1, PlayerPermissions: packet.PermissionLevelMember, Layers: []protocol.AbilityLayer{{
		Type: protocol.AbilityLayerTypeBase, Abilities: protocol.AbilityCount - 1, Values: values,
	}}}
}

func TestAbilitiesBecomeAdventureSettings(t *testing.T) {
	interact := uint32(protocol.AbilityDoorsAndSwitches | protocol.AbilityOpenContainers | protocol.AbilityAttackPlayers | protocol.AbilityAttackMobs)
	edit := uint32(protocol.AbilityBuild | protocol.AbilityMine)
	survival := adventureFrom(abilities(edit | interact))
	creative := adventureFrom(abilities(edit | interact | protocol.AbilityMayFly | protocol.AbilityInstantBuild | protocol.AbilityInvulnerable))
	spectator := adventureFrom(abilities(protocol.AbilityMayFly | protocol.AbilityFlying | protocol.AbilityNoClip | protocol.AbilityInvulnerable))

	if survival.Flags&(packet.AdventureFlagWorldImmutable|packet.AdventureFlagAllowFlight) != 0 {
		t.Errorf("survival flags %b", survival.Flags)
	}
	if survival.ActionPermissions&(packet.ActionPermissionBuild|packet.ActionPermissionMine) != packet.ActionPermissionBuild|packet.ActionPermissionMine {
		t.Errorf("survival cannot build: %b", survival.ActionPermissions)
	}
	if creative.Flags&packet.AdventureFlagAllowFlight == 0 || creative.Flags&packet.AdventureFlagFlying != 0 {
		t.Errorf("creative flags %b", creative.Flags)
	}
	want := uint32(packet.AdventureFlagWorldImmutable | packet.AdventureFlagNoClip | packet.AdventureFlagFlying | packet.AdventureFlagAllowFlight)
	if spectator.Flags&want != want || spectator.ActionPermissions&packet.ActionPermissionBuild != 0 {
		t.Errorf("spectator flags %b, actions %b", spectator.Flags, spectator.ActionPermissions)
	}
	if survival.PlayerUniqueID != 1 || survival.PermissionLevel != packet.PermissionLevelMember {
		t.Errorf("identity lost: %+v", survival)
	}
}

func TestUpdateAbilitiesReachOldClients(t *testing.T) {
	out := New(527).ConvertFromLatest(&packet.UpdateAbilities{AbilityData: abilities(protocol.AbilityMayFly)}, nil)
	if len(out) != 1 {
		t.Fatalf("UpdateAbilities became %d packets", len(out))
	}
	a, ok := out[0].(*packet.AdventureSettings)
	if !ok || a.Flags&packet.AdventureFlagAllowFlight == 0 {
		t.Fatalf("UpdateAbilities became %#v", out[0])
	}
}
