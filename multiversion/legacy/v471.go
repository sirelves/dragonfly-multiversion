package legacy

import (
	"github.com/df-mc/dragonfly/multiversion/v630"
	"github.com/df-mc/dragonfly/multiversion/v786"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// 1.17.40 (471) and older → 1.18.0 (475):
//   - Worlds spanned y 0 to 255: chunks are rewritten by legacyChunk, and
//     block updates outside that range are not sent.
//   - StartGame had no block state checksum (gated in v766).
var delta471 = delta{
	proto: 471,
	fromLatest: func(_ uint32, pk packet.Packet) []packet.Packet {
		switch pk := pk.(type) {
		case *v630.LevelChunk:
			if pk.CacheEnabled || pk.SubChunkCount >= v786.SubChunkRequestModeLimited786 {
				return keep(pk) // Dragonfly sends whole chunks without the blob cache to these clients
			}
			raw, n, err := legacyChunk(pk.RawPayload, pk.SubChunkCount, pk.Dimension)
			if err != nil {
				return nil
			}
			cp := *pk
			cp.RawPayload, cp.SubChunkCount = raw, n
			return []packet.Packet{&cp}
		case *v786.UpdateBlock:
			return inWorld(pk, pk.Position)
		case *v786.UpdateBlockSynced:
			return inWorld(pk, pk.Position)
		case *v786.BlockActorData:
			return inWorld(pk, pk.Position)
		}
		return keep(pk)
	},
}

// inWorld keeps pk only if pos lies in the y 0–255 world of old clients.
func inWorld(pk packet.Packet, pos protocol.BlockPos) []packet.Packet {
	if pos[1] < 0 || pos[1] > 255 {
		return nil
	}
	return keep(pk)
}
