package legacy

import (
	"github.com/df-mc/dragonfly/multiversion/v686"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// 1.16.210 (428) and older → 1.16.220 (431), CloudburstMC v431:
//   - Items were written inline in the old form (v786 does that below 431 for
//     every Item and ItemInstance). Only InventoryContent and InventorySlot
//     carried a stack network ID, as a varint in front of the item.
//   - InventoryTransaction flagged whether its actions carry stack network IDs
//     (v786).
//   - Packets were encrypted with AES-CFB8 (handled by the listener).
var delta428 = delta{
	proto: 428,
	packets: func(_ uint32, pool packet.Pool, _ bool) {
		if _, ok := pool[packet.IDInventoryContent]; ok {
			pool[packet.IDInventoryContent] = func() packet.Packet { return &InventoryContent428{} }
		}
		if _, ok := pool[packet.IDInventorySlot]; ok {
			pool[packet.IDInventorySlot] = func() packet.Packet { return &InventorySlot428{} }
		}
	},
	fromLatest: func(_ uint32, pk packet.Packet) []packet.Packet {
		switch pk := pk.(type) {
		case *v686.InventoryContent:
			return []packet.Packet{&InventoryContent428{WindowID: pk.WindowID, Content: pk.Content}}
		case *v686.InventorySlot:
			return []packet.Packet{&InventorySlot428{WindowID: pk.WindowID, Slot: pk.Slot, NewItem: pk.NewItem}}
		}
		return keep(pk)
	},
	toLatest: func(_ uint32, pk packet.Packet) []packet.Packet {
		switch pk := pk.(type) {
		case *InventoryContent428:
			return []packet.Packet{&v686.InventoryContent{WindowID: pk.WindowID, Content: pk.Content}}
		case *InventorySlot428:
			return []packet.Packet{&v686.InventorySlot{WindowID: pk.WindowID, Slot: pk.Slot, NewItem: pk.NewItem}}
		}
		return keep(pk)
	},
}

// netItem428 reads or writes a stack network ID followed by an old-form item.
func netItem428(io protocol.IO, x *protocol.ItemInstance) {
	id := x.StackNetworkID
	io.Varint32(&id)
	io.ItemInstance(x) // below 431 v786 handles this as a bare old-form item
	x.StackNetworkID = id
}

// InventoryContent428 is InventoryContent before 1.16.220.
type InventoryContent428 struct {
	WindowID uint32
	Content  []protocol.ItemInstance
}

func (*InventoryContent428) ID() uint32 { return packet.IDInventoryContent }

func (pk *InventoryContent428) Marshal(io protocol.IO) {
	io.Varuint32(&pk.WindowID)
	protocol.FuncIOSlice(io, &pk.Content, netItem428)
}

// InventorySlot428 is InventorySlot before 1.16.220.
type InventorySlot428 struct {
	WindowID uint32
	Slot     uint32
	NewItem  protocol.ItemInstance
}

func (*InventorySlot428) ID() uint32 { return packet.IDInventorySlot }

func (pk *InventorySlot428) Marshal(io protocol.IO) {
	io.Varuint32(&pk.WindowID)
	io.Varuint32(&pk.Slot)
	netItem428(io, &pk.NewItem)
}
