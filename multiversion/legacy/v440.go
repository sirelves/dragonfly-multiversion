package legacy

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// 1.17.0 (440) → 1.17.10 (448), CloudburstMC v448:
//   - SetTitle gained the XUID and platform online ID, ResourcePacksInfo the
//     forcing server packs flag (gated in v686).
//   - NPCRequest had no scene name.
var delta440 = delta{
	proto: 440,
	packets: func(_ uint32, pool packet.Pool, listener bool) {
		if listener {
			pool[packet.IDNPCRequest] = func() packet.Packet { return &NPCRequest440{} }
		}
	},
	fromLatest: func(_ uint32, pk packet.Packet) []packet.Packet {
		if pk, ok := pk.(*packet.NPCRequest); ok {
			return []packet.Packet{&NPCRequest440{EntityRuntimeID: pk.EntityRuntimeID, RequestType: pk.RequestType, CommandString: pk.CommandString, ActionType: pk.ActionType}}
		}
		return keep(pk)
	},
	toLatest: func(_ uint32, pk packet.Packet) []packet.Packet {
		if pk, ok := pk.(*NPCRequest440); ok {
			return []packet.Packet{&packet.NPCRequest{EntityRuntimeID: pk.EntityRuntimeID, RequestType: pk.RequestType, CommandString: pk.CommandString, ActionType: pk.ActionType}}
		}
		return keep(pk)
	},
}

// NPCRequest440 is NPCRequest before 1.17.10.
type NPCRequest440 struct {
	EntityRuntimeID uint64
	RequestType     byte
	CommandString   string
	ActionType      byte
}

func (*NPCRequest440) ID() uint32 { return packet.IDNPCRequest }

func (pk *NPCRequest440) Marshal(io protocol.IO) {
	io.ActorRuntimeID(&pk.EntityRuntimeID)
	io.Uint8(&pk.RequestType)
	io.String(&pk.CommandString)
	io.Uint8(&pk.ActionType)
}
