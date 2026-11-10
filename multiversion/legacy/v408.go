package legacy

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"io"
	"sync"

	"github.com/df-mc/dragonfly/multiversion/v766"
	"github.com/df-mc/dragonfly/multiversion/v786"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// 1.16.20 (408) and 1.16.0 (407) → 1.16.100 (419), CloudburstMC v419:
//   - StartGame sent the whole block palette as an NBT list, which also fixed
//     the runtime IDs, items as name and legacy ID only, a server
//     authoritative movement bool and no experiments; ResourcePackStack had an
//     experimental flag in their place (gated in v766 and v786).
//   - MovePlayer, UpdateAttributes, SetActorData and PlayerAuthInput had no
//     tick (PlayerAuthInput no delta either), ContainerClose no server-side
//     flag, and ItemStackResponse a success bool where Status 0 means OK
//     (gated in v630, v712 and v786).
//   - MoveActorDelta carried delta varints. Dragonfly never sends it; it is
//     dropped rather than rebuilt.
var delta408 = delta{
	proto: 408,
	fromLatest: func(_ uint32, pk packet.Packet) []packet.Packet {
		switch pk := pk.(type) {
		case *v766.StartGame:
			cp := *pk
			cp.LegacyBlockPalette = blockPalette407()
			return []packet.Packet{&cp}
		case *packet.MoveActorDelta, *v786.MoveActorDelta:
			return nil
		}
		return keep(pk)
	},
	toLatest: func(_ uint32, pk packet.Packet) []packet.Packet {
		// Items came with StartGame and nothing followed it. Clients built on
		// the current gophertunnel (the test dialers) wait for an ItemRegistry,
		// so it is made from the StartGame items. Servers never see StartGame.
		if sg, ok := pk.(*v766.StartGame); ok {
			reg := &packet.ItemRegistry{Items: make([]protocol.ItemEntry, len(sg.Items))}
			for i, it := range sg.Items {
				reg.Items[i] = protocol.ItemEntry{Name: it.Name, RuntimeID: it.RuntimeID}
			}
			return []packet.Packet{sg, reg}
		}
		return keep(pk)
	},
}

// block_palette_407.nbt.gz is canonical_block_states.nbt of 1.16.0, which is
// the same for 1.16.20: the list StartGame carries, in runtime ID order.
//
//go:embed data/block_palette_407.nbt.gz
var blockPalette407gz []byte

var blockPalette407 = sync.OnceValue(func() []byte {
	r, err := gzip.NewReader(bytes.NewReader(blockPalette407gz))
	if err != nil {
		panic("legacy: " + err.Error())
	}
	data, err := io.ReadAll(r)
	if err != nil {
		panic("legacy: " + err.Error())
	}
	return data
})
