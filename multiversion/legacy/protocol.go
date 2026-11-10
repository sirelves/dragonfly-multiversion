// Package legacy implements the Bedrock protocols older than 1.20.0, from
// 1.16.0 (407) to 1.19.80 (582).
//
// Each client still goes through v630.Protocol with its own protocol number,
// so block palettes, item registries, biomes and sounds are looked up for the
// right version. What v630 produces is the 1.20.0 wire format; a chain of
// deltas then walks it down one release at a time. A delta knows only how its
// release differs from the next newer one, which keeps every step small and
// checkable against the gophertunnel tag that shipped it.
package legacy

import (
	"github.com/df-mc/dragonfly/multiversion/v630"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// Protocol is a pre-1.20 Bedrock protocol.
type Protocol struct{ id int32 }

// New returns the Protocol for id. It panics if id is not a supported legacy
// protocol.
func New(id int32) Protocol {
	if _, ok := versions[id]; !ok {
		panic("legacy: unsupported protocol")
	}
	return Protocol{id: id}
}

// Protocols returns every supported legacy protocol, newest first.
func Protocols() []minecraft.Protocol {
	out := make([]minecraft.Protocol, 0, len(order))
	for _, id := range order {
		out = append(out, Protocol{id: id})
	}
	return out
}

// order lists the supported protocols, newest first.
var order = []int32{582, 575, 568, 567, 560, 557, 554, 545, 544, 534, 527}

var versions = map[int32]string{
	582: "1.19.80",
	575: "1.19.70",
	568: "1.19.63",
	567: "1.19.60",
	560: "1.19.50",
	557: "1.19.40",
	554: "1.19.30",
	545: "1.19.21",
	544: "1.19.20",
	534: "1.19.10",
	527: "1.19.0",
}

func (p Protocol) ID() int32   { return p.id }
func (p Protocol) Ver() string { return versions[p.id] }

func (p Protocol) base() v630.Protocol { return v630.Protocol{Version: p.id} }

func (p Protocol) Packets(listener bool) packet.Pool {
	pool := p.base().Packets(listener)
	for _, d := range p.deltas() {
		if d.packets != nil {
			d.packets(uint32(p.id), pool, listener)
		}
	}
	for id := range pool {
		if _, ok := knownSets[p.id][id]; !ok && !handshake(id) {
			delete(pool, id) // a packet this client cannot send or receive
		}
	}
	return pool
}

func (p Protocol) NewReader(r minecraft.ByteReader, shieldID int32, enableLimits bool) protocol.IO {
	return p.base().NewReader(r, shieldID, enableLimits)
}

func (p Protocol) NewWriter(w minecraft.ByteWriter, shieldID int32) protocol.IO {
	return p.base().NewWriter(w, shieldID)
}

// ConvertToLatest walks a packet up from the client version to 1.20.0 and
// hands it to v630.
func (p Protocol) ConvertToLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	pks := []packet.Packet{pk}
	ds := p.deltas()
	for i := len(ds) - 1; i >= 0; i-- {
		if ds[i].toLatest != nil {
			pks = apply(pks, uint32(p.id), ds[i].toLatest)
		}
	}
	out := make([]packet.Packet, 0, len(pks))
	for _, pk := range pks {
		out = append(out, p.base().ConvertToLatest(pk, conn)...)
	}
	return out
}

// ConvertFromLatest lets v630 produce the 1.20.0 form and walks it down to the
// client version.
func (p Protocol) ConvertFromLatest(pk packet.Packet, conn *minecraft.Conn) []packet.Packet {
	pks := p.base().ConvertFromLatest(pk, conn)
	for _, d := range p.deltas() {
		if d.fromLatest != nil {
			pks = apply(pks, uint32(p.id), d.fromLatest)
		}
	}
	return p.knownOnly(pks)
}

// knownOnly drops packets the client of p has no decoder for: newer server
// features have nothing to say to an older client.
func (p Protocol) knownOnly(pks []packet.Packet) []packet.Packet {
	known := knownSets[p.id]
	out := pks[:0]
	for _, pk := range pks {
		if _, ok := known[pk.ID()]; ok || handshake(pk.ID()) {
			out = append(out, pk)
		}
	}
	return out
}

// handshake reports whether id is one of the packets the RakNet listener
// rewrites for clients before 1.19.30, which must reach it intact.
func handshake(id uint32) bool {
	return id == packet.IDRequestNetworkSettings || id == packet.IDNetworkSettings
}

var knownSets = func() map[int32]map[uint32]struct{} {
	m := make(map[int32]map[uint32]struct{}, len(knownIDs))
	for proto, ids := range knownIDs {
		set := make(map[uint32]struct{}, len(ids))
		for _, id := range ids {
			set[id] = struct{}{}
		}
		m[proto] = set
	}
	return m
}()

func apply(pks []packet.Packet, proto uint32, f func(proto uint32, pk packet.Packet) []packet.Packet) []packet.Packet {
	out := pks[:0:0]
	for _, pk := range pks {
		out = append(out, f(proto, pk)...)
	}
	return out
}

// delta describes how one release differs from the next newer one. fromLatest
// receives packets in the newer release's form and returns them in this
// release's form; toLatest does the opposite. Both return the packet unchanged
// when it is not theirs, and nil to drop it.
type delta struct {
	proto      int32
	packets    func(proto uint32, pool packet.Pool, listener bool)
	fromLatest func(proto uint32, pk packet.Packet) []packet.Packet
	toLatest   func(proto uint32, pk packet.Packet) []packet.Packet
}

// deltas is ordered newest first.
var deltas = []delta{delta582, delta575, delta567, delta560, delta557, delta554, delta545, delta544, delta534, delta527}

// deltas returns the deltas between 1.20.0 and p, newest first.
func (p Protocol) deltas() []delta {
	n := 0
	for n < len(deltas) && deltas[n].proto >= p.id {
		n++
	}
	return deltas[:n]
}

func keep(pk packet.Packet) []packet.Packet { return []packet.Packet{pk} }
