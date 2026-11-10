package legacy

import (
	"bytes"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// 1.19.10 (534) → 1.19.20 (544), CloudburstMC v544:
//   - StartGame gained PersonaDisabled, CustomSkinsDisabled,
//     ChatRestrictionLevel, DisablePlayerInteractions and ClientSideGeneration
//     (gated in v766); attributes gained modifiers (gated in v712) and
//     ClientBoundMapItemData its origin (gated in v786).
//   - NetworkChunkPublisherUpdate gained the saved chunk list, MapInfoRequest
//     the client pixels, and ModalFormResponse turned its JSON body into an
//     optional next to an optional cancel reason. Before that a closed form
//     answered "null".
var delta534 = delta{
	proto: 534,
	packets: func(_ uint32, pool packet.Pool, listener bool) {
		if listener {
			pool[packet.IDModalFormResponse] = func() packet.Packet { return &ModalFormResponse534{} }
			pool[packet.IDMapInfoRequest] = func() packet.Packet { return &MapInfoRequest534{} }
		} else {
			pool[packet.IDNetworkChunkPublisherUpdate] = func() packet.Packet { return &NetworkChunkPublisherUpdate534{} }
		}
	},
	fromLatest: func(_ uint32, pk packet.Packet) []packet.Packet {
		switch pk := pk.(type) {
		case *packet.NetworkChunkPublisherUpdate:
			return []packet.Packet{&NetworkChunkPublisherUpdate534{Position: pk.Position, Radius: pk.Radius}}
		case *packet.ModalFormResponse:
			out := &ModalFormResponse534{FormID: pk.FormID, ResponseData: []byte("null")}
			if data, ok := pk.ResponseData.Value(); ok {
				out.ResponseData = data
			}
			return []packet.Packet{out}
		case *packet.MapInfoRequest:
			return []packet.Packet{&MapInfoRequest534{MapID: pk.MapID}}
		}
		return keep(pk)
	},
	toLatest: func(_ uint32, pk packet.Packet) []packet.Packet {
		switch pk := pk.(type) {
		case *ModalFormResponse534:
			out := &packet.ModalFormResponse{FormID: pk.FormID}
			if bytes.Equal(bytes.TrimSpace(pk.ResponseData), []byte("null")) {
				out.CancelReason = protocol.Option[uint8](packet.ModalFormCancelReasonUserClosed)
			} else {
				out.ResponseData = protocol.Option(pk.ResponseData)
			}
			return []packet.Packet{out}
		case *MapInfoRequest534:
			return []packet.Packet{&packet.MapInfoRequest{MapID: pk.MapID}}
		case *NetworkChunkPublisherUpdate534:
			return []packet.Packet{&packet.NetworkChunkPublisherUpdate{Position: pk.Position, Radius: pk.Radius}}
		}
		return keep(pk)
	},
}

// ModalFormResponse534 is ModalFormResponse before 1.19.20.
type ModalFormResponse534 struct {
	FormID       uint32
	ResponseData []byte
}

func (*ModalFormResponse534) ID() uint32 { return packet.IDModalFormResponse }

func (pk *ModalFormResponse534) Marshal(io protocol.IO) {
	io.Varuint32(&pk.FormID)
	io.ByteSlice(&pk.ResponseData)
}

// MapInfoRequest534 is MapInfoRequest before 1.19.20.
type MapInfoRequest534 struct{ MapID int64 }

func (*MapInfoRequest534) ID() uint32 { return packet.IDMapInfoRequest }

func (pk *MapInfoRequest534) Marshal(io protocol.IO) { io.Varint64(&pk.MapID) }

// NetworkChunkPublisherUpdate534 is NetworkChunkPublisherUpdate before 1.19.20.
type NetworkChunkPublisherUpdate534 struct {
	Position protocol.BlockPos
	Radius   uint32
}

func (*NetworkChunkPublisherUpdate534) ID() uint32 { return packet.IDNetworkChunkPublisherUpdate }

func (pk *NetworkChunkPublisherUpdate534) Marshal(io protocol.IO) {
	io.BlockPos(&pk.Position)
	io.Varuint32(&pk.Radius)
}
