package legacy

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// 1.18.10 (486) → 1.18.30 (503), CloudburstMC v503:
//   - StartGame's world seed was a varint32, AddPlayer had no game type, volume
//     entities had no bounds or dimension (gated in v766 and v786).
//   - SpawnParticleEffect had no MoLang variables.
var delta486 = delta{
	proto: 486,
	fromLatest: func(_ uint32, pk packet.Packet) []packet.Packet {
		if pk, ok := pk.(*packet.SpawnParticleEffect); ok {
			return []packet.Packet{&SpawnParticleEffect486{Dimension: pk.Dimension, EntityUniqueID: pk.EntityUniqueID, Position: pk.Position, ParticleName: pk.ParticleName}}
		}
		return keep(pk)
	},
}

// SpawnParticleEffect486 is SpawnParticleEffect before 1.18.30.
type SpawnParticleEffect486 struct {
	Dimension      byte
	EntityUniqueID int64
	Position       mgl32.Vec3
	ParticleName   string
}

func (*SpawnParticleEffect486) ID() uint32 { return packet.IDSpawnParticleEffect }

func (pk *SpawnParticleEffect486) Marshal(io protocol.IO) {
	io.Uint8(&pk.Dimension)
	io.Varint64(&pk.EntityUniqueID)
	io.Vec3(&pk.Position)
	io.String(&pk.ParticleName)
}
