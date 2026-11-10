package legacy

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// 1.17.10 (448) → 1.17.30 (465), CloudburstMC v465:
//   - StartGame gained the education shared resource URI, skins the geometry
//     engine version and primary user flag, AddVolumeEntity its engine
//     version, HurtArmour the slot mask and CraftingData the material reducers
//     (gated in v630, v766 and v786).
//   - ActorPickRequest had no WithData flag. EducationSettings and
//     PhotoTransfer, which Dragonfly never sends, had fewer fields; they are
//     dropped rather than rebuilt.
var delta448 = delta{
	proto: 448,
	packets: func(_ uint32, pool packet.Pool, listener bool) {
		if listener {
			pool[packet.IDActorPickRequest] = func() packet.Packet { return &ActorPickRequest448{} }
		}
	},
	fromLatest: func(_ uint32, pk packet.Packet) []packet.Packet {
		switch pk := pk.(type) {
		case *packet.EducationSettings, *packet.PhotoTransfer:
			return nil
		case *packet.ActorPickRequest:
			return []packet.Packet{&ActorPickRequest448{EntityUniqueID: pk.EntityUniqueID, HotBarSlot: pk.HotBarSlot}}
		}
		return keep(pk)
	},
	toLatest: func(_ uint32, pk packet.Packet) []packet.Packet {
		if pk, ok := pk.(*ActorPickRequest448); ok {
			return []packet.Packet{&packet.ActorPickRequest{EntityUniqueID: pk.EntityUniqueID, HotBarSlot: pk.HotBarSlot}}
		}
		return keep(pk)
	},
}

// ActorPickRequest448 is ActorPickRequest before 1.17.30.
type ActorPickRequest448 struct {
	EntityUniqueID int64
	HotBarSlot     byte
}

func (*ActorPickRequest448) ID() uint32 { return packet.IDActorPickRequest }

func (pk *ActorPickRequest448) Marshal(io protocol.IO) {
	io.ActorUniqueIDInt64(&pk.EntityUniqueID)
	io.Uint8(&pk.HotBarSlot)
}
