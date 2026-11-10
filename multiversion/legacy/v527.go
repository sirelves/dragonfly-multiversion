package legacy

import (
	"github.com/df-mc/dragonfly/multiversion/v766"
	"github.com/df-mc/dragonfly/multiversion/v786"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// 1.19.0 (527) → 1.19.10 (534), CloudburstMC v534:
//   - Abilities did not exist: UpdateAbilities becomes AdventureSettings, and
//     AddPlayer carried the adventure settings itself, along with the entity
//     unique ID.
//   - StartGame had no editor flag (gated in v766), AddActor no body yaw
//     (gated in v786).
var delta527 = delta{
	proto: 527,
	fromLatest: func(_ uint32, pk packet.Packet) []packet.Packet {
		switch pk := pk.(type) {
		case *packet.UpdateAbilities:
			a := adventureFrom(pk.AbilityData)
			return []packet.Packet{&a}
		case *v766.UpdateAbilities:
			a := adventureFrom(abilityData766(pk.AbilityData))
			return []packet.Packet{&a}
		case *packet.UpdateAdventureSettings:
			return nil // folded into the abilities, which Dragonfly always sends
		case *v786.AddPlayer:
			return []packet.Packet{&AddPlayer527{AddPlayer: pk, Adventure: adventureFrom(pk.AbilityData)}}
		}
		return keep(pk)
	},
}

func abilityData766(d v766.AbilityData) protocol.AbilityData {
	out := protocol.AbilityData{EntityUniqueID: d.EntityUniqueID, PlayerPermissions: d.PlayerPermissions, CommandPermissions: d.CommandPermissions}
	for _, l := range d.Layers {
		out.Layers = append(out.Layers, protocol.AbilityLayer{Type: l.Type, Abilities: l.Abilities, Values: l.Values, FlySpeed: l.FlySpeed, WalkSpeed: l.WalkSpeed})
	}
	return out
}

// adventureFrom flattens ability layers into the adventure settings of the
// clients that predate abilities.
func adventureFrom(d protocol.AbilityData) packet.AdventureSettings {
	var values uint32
	for _, l := range d.Layers {
		values = values&^l.Abilities | l.Values&l.Abilities
	}
	has := func(a uint32) bool { return values&a != 0 }
	flags := uint32(packet.AdventureFlagAutoJump)
	set := func(cond bool, f uint32) {
		if cond {
			flags |= f
		}
	}
	set(!has(protocol.AbilityBuild) && !has(protocol.AbilityMine), packet.AdventureFlagWorldImmutable)
	set(!has(protocol.AbilityAttackMobs), packet.AdventureSettingsFlagsNoPvM)
	set(!has(protocol.AbilityAttackPlayers), packet.AdventureSettingsFlagsNoMvP)
	set(has(protocol.AbilityMayFly), packet.AdventureFlagAllowFlight)
	set(has(protocol.AbilityNoClip), packet.AdventureFlagNoClip)
	set(has(protocol.AbilityWorldBuilder), packet.AdventureFlagWorldBuilder)
	set(has(protocol.AbilityFlying), packet.AdventureFlagFlying)
	set(has(protocol.AbilityMuted), packet.AdventureFlagMuted)

	var actions uint32
	for ability, action := range map[uint32]uint32{
		protocol.AbilityMine:             packet.ActionPermissionMine,
		protocol.AbilityDoorsAndSwitches: packet.ActionPermissionDoorsAndSwitches,
		protocol.AbilityOpenContainers:   packet.ActionPermissionOpenContainers,
		protocol.AbilityAttackPlayers:    packet.ActionPermissionAttackPlayers,
		protocol.AbilityAttackMobs:       packet.ActionPermissionAttackMobs,
		protocol.AbilityOperatorCommands: packet.ActionPermissionOperator,
		protocol.AbilityTeleport:         packet.ActionPermissionTeleport,
		protocol.AbilityBuild:            packet.ActionPermissionBuild,
	} {
		if has(ability) {
			actions |= action
		}
	}
	return packet.AdventureSettings{
		Flags:                  flags,
		CommandPermissionLevel: uint32(d.CommandPermissions),
		ActionPermissions:      actions,
		PermissionLevel:        uint32(d.PlayerPermissions),
		PlayerUniqueID:         d.EntityUniqueID,
	}
}

// AddPlayer527 is AddPlayer before 1.19.10.
type AddPlayer527 struct {
	*v786.AddPlayer
	Adventure packet.AdventureSettings
}

func (*AddPlayer527) ID() uint32 { return packet.IDAddPlayer }

func (pk *AddPlayer527) Marshal(io protocol.IO) {
	io.UUID(&pk.UUID)
	io.String(&pk.Username)
	io.Varint64(&pk.Adventure.PlayerUniqueID)
	io.Varuint64(&pk.EntityRuntimeID)
	io.String(&pk.PlatformChatID)
	io.Vec3(&pk.Position)
	io.Vec3(&pk.Velocity)
	io.Float32(&pk.Pitch)
	io.Float32(&pk.Yaw)
	io.Float32(&pk.HeadYaw)
	io.ItemInstance(&pk.HeldItem)
	if p := v786.ProtoOf(io); p == 0 || p >= 503 { // added in 503
		io.Varint32(&pk.GameType)
	}
	io.EntityMetadata(&pk.EntityMetadata)
	io.Varuint32(&pk.Adventure.Flags)
	io.Varuint32(&pk.Adventure.CommandPermissionLevel)
	io.Varuint32(&pk.Adventure.ActionPermissions)
	io.Varuint32(&pk.Adventure.PermissionLevel)
	io.Varuint32(&pk.Adventure.CustomStoredPermissions)
	io.Int64(&pk.Adventure.PlayerUniqueID)
	v786.EntityLinks786(io, &pk.EntityLinks)
	io.String(&pk.DeviceID)
	io.Int32(&pk.BuildPlatform)
}
