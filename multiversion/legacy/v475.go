package legacy

import (
	"github.com/df-mc/dragonfly/multiversion/v786"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// 1.18.0 (475) → 1.18.10 (486), CloudburstMC v486:
//   - Sub-chunk requests named one sub-chunk each: SubChunk carried a single
//     absolute position and SubChunkRequest asked for one. The limited request
//     mode of LevelChunk did not exist (mapped to limitless in v630).
//     Dragonfly sends whole chunks, so this only keeps the wire honest.
//   - AddVolumeEntity had no encoding or instance identifier (gated in v786).
var delta475 = delta{
	proto: 475,
	packets: func(_ uint32, pool packet.Pool, listener bool) {
		if listener {
			pool[packet.IDSubChunkRequest] = func() packet.Packet { return &SubChunkRequest475{} }
		} else {
			pool[packet.IDSubChunk] = func() packet.Packet { return &SubChunk475{} }
		}
	},
	fromLatest: func(_ uint32, pk packet.Packet) []packet.Packet {
		switch r := pk.(type) {
		case *packet.SubChunkRequest:
			return requests475(r.Dimension, r.Position, r.Offsets)
		case *v786.SubChunkRequest:
			return requests475(r.Dimension, r.Position, r.Offsets)
		}
		sc, ok := pk.(*v786.SubChunk)
		if !ok {
			return keep(pk)
		}
		out := make([]packet.Packet, 0, len(sc.SubChunkEntries))
		for _, e := range sc.SubChunkEntries {
			p := &SubChunk475{Dimension: sc.Dimension, CacheEnabled: sc.CacheEnabled, BlobHash: e.BlobHash, HeightMapType: e.HeightMapType,
				X: sc.Position[0] + int32(e.Offset[0]), Y: sc.Position[1] + int32(e.Offset[1]), Z: sc.Position[2] + int32(e.Offset[2]),
				Data: e.RawPayload, Result: int32(e.Result)}
			if e.Result == protocol.SubChunkResultSuccessAllAir {
				p.Result, p.Data = protocol.SubChunkResultSuccess, []byte{8, 0} // an empty sub-chunk
			}
			if p.HeightMapType == protocol.HeightMapDataHasData {
				p.HeightMapData = make([]byte, 256)
				for i := range p.HeightMapData {
					if i < len(e.HeightMapData) {
						p.HeightMapData[i] = byte(e.HeightMapData[i])
					}
				}
			}
			out = append(out, p)
		}
		return out
	},
	toLatest: func(_ uint32, pk packet.Packet) []packet.Packet {
		if r, ok := pk.(*SubChunkRequest475); ok {
			return []packet.Packet{&packet.SubChunkRequest{Dimension: r.Dimension, Position: protocol.SubChunkPos{r.X, r.Y, r.Z}, Offsets: []protocol.SubChunkOffset{{}}}}
		}
		return keep(pk)
	},
}

func requests475(dim int32, pos protocol.SubChunkPos, offsets []protocol.SubChunkOffset) []packet.Packet {
	if len(offsets) == 0 {
		offsets = []protocol.SubChunkOffset{{}}
	}
	out := make([]packet.Packet, len(offsets))
	for i, o := range offsets {
		out[i] = &SubChunkRequest475{Dimension: dim, X: pos[0] + int32(o[0]), Y: pos[1] + int32(o[1]), Z: pos[2] + int32(o[2])}
	}
	return out
}

// SubChunk475 is SubChunk before 1.18.10: one sub-chunk per packet.
type SubChunk475 struct {
	Dimension     int32
	X, Y, Z       int32
	Data          []byte
	Result        int32
	HeightMapType byte
	HeightMapData []byte
	CacheEnabled  bool
	BlobHash      uint64
}

func (*SubChunk475) ID() uint32 { return packet.IDSubChunk }

func (pk *SubChunk475) Marshal(io protocol.IO) {
	io.Varint32(&pk.Dimension)
	io.Varint32(&pk.X)
	io.Varint32(&pk.Y)
	io.Varint32(&pk.Z)
	io.ByteSlice(&pk.Data)
	io.Varint32(&pk.Result)
	io.Uint8(&pk.HeightMapType)
	if pk.HeightMapType == protocol.HeightMapDataHasData {
		if len(pk.HeightMapData) != 256 {
			pk.HeightMapData = make([]byte, 256)
		}
		for i := range pk.HeightMapData {
			io.Uint8(&pk.HeightMapData[i])
		}
	}
	io.Bool(&pk.CacheEnabled)
	if pk.CacheEnabled {
		io.Uint64(&pk.BlobHash)
	}
}

// SubChunkRequest475 is SubChunkRequest before 1.18.10: one position.
type SubChunkRequest475 struct {
	Dimension int32
	X, Y, Z   int32
}

func (*SubChunkRequest475) ID() uint32 { return packet.IDSubChunkRequest }

func (pk *SubChunkRequest475) Marshal(io protocol.IO) {
	io.Varint32(&pk.Dimension)
	io.Varint32(&pk.X)
	io.Varint32(&pk.Y)
	io.Varint32(&pk.Z)
}
