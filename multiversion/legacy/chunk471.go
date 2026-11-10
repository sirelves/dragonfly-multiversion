package legacy

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// Clients before 1.18.0 know worlds from y 0 to 255 only: at most 16
// sub-chunks counted from y 0, and one biome per column (256 bytes) instead of
// a paletted biome storage per sub-chunk. legacyChunk rewrites a LevelChunk
// payload of Dragonfly's network encoding (version 8 sub-chunks for the whole
// dimension range, a biome storage per sub-chunk, then the border block byte
// and block entity NBT) into that form. Blocks and block entities outside
// 0–255 are not sent; the column biome is the one at y 64.
func legacyChunk(payload []byte, count uint32, dimension int32) ([]byte, uint32, error) {
	dim, ok := world.DimensionByID(int(dimension))
	if !ok {
		return nil, 0, fmt.Errorf("unknown dimension %d", dimension)
	}
	r := dim.Range()
	minIndex := r.Min() >> 4
	storages := (r.Max() - r.Min() + 1) >> 4 // Range.Height is Max-Min
	in := bytes.NewBuffer(payload)

	subs := make([][]byte, count)
	for i := range subs {
		start := in.Len()
		if err := skipSubChunk(in); err != nil {
			return nil, 0, fmt.Errorf("sub-chunk %d: %w", i, err)
		}
		subs[i] = payload[len(payload)-start : len(payload)-in.Len()]
	}
	biomeAt := 4 - minIndex // the storage holding y 64..79
	var column []byte
	var previous *biomeStorage
	for i := 0; i < storages; i++ {
		s, err := readBiomeStorage(in, previous)
		if err != nil {
			return nil, 0, fmt.Errorf("biome storage %d: %w", i, err)
		}
		if i == biomeAt || (column == nil && i == storages-1) {
			column = s.column(0)
		}
		previous = s
	}
	if column == nil {
		column = make([]byte, 256)
	}
	var border byte
	if in.Len() > 0 {
		border, _ = in.ReadByte()
	}
	if border != 0 {
		return nil, 0, errors.New("border blocks are not supported")
	}

	out := new(bytes.Buffer)
	last := -1
	for i := range subs {
		if y := minIndex + i; y >= 0 && y < 16 && !emptySubChunk(subs[i]) {
			last = y
		}
	}
	for y := 0; y <= last; y++ {
		if i := y - minIndex; i >= 0 && i < len(subs) {
			out.Write(subs[i])
		} else {
			out.Write([]byte{8, 0})
		}
	}
	out.Write(column)
	out.WriteByte(0)

	dec := nbt.NewDecoderWithEncoding(in, nbt.NetworkLittleEndian)
	enc := nbt.NewEncoderWithEncoding(out, nbt.NetworkLittleEndian)
	for in.Len() > 0 {
		var be map[string]any
		if err := dec.Decode(&be); err != nil {
			return nil, 0, fmt.Errorf("block entity: %w", err)
		}
		if y, ok := be["y"].(int32); ok && (y < 0 || y > 255) {
			continue
		}
		if err := enc.Encode(be); err != nil {
			return nil, 0, err
		}
	}
	return out.Bytes(), uint32(last + 1), nil
}

func emptySubChunk(sub []byte) bool { return len(sub) < 2 || sub[1] == 0 }

// skipSubChunk consumes one network sub-chunk of version 8 or 9.
func skipSubChunk(in *bytes.Buffer) error {
	ver, err := in.ReadByte()
	if err != nil {
		return err
	}
	if ver != 8 && ver != 9 {
		return fmt.Errorf("sub-chunk version %d", ver)
	}
	n, err := in.ReadByte()
	if err != nil {
		return err
	}
	if ver == 9 {
		if _, err := in.ReadByte(); err != nil {
			return err
		}
	}
	for i := byte(0); i < n; i++ {
		if _, err := readStorage(in, nil); err != nil {
			return err
		}
	}
	return nil
}

type biomeStorage struct {
	bits    byte
	words   []byte
	palette []int32
}

func readBiomeStorage(in *bytes.Buffer, previous *biomeStorage) (*biomeStorage, error) {
	return readStorage(in, previous)
}

// readStorage reads a network paletted storage. Header 0xff repeats the
// previous storage, which only biome storages use.
func readStorage(in *bytes.Buffer, previous *biomeStorage) (*biomeStorage, error) {
	header, err := in.ReadByte()
	if err != nil {
		return nil, err
	}
	bits := header >> 1
	if bits == 0x7f {
		if previous == nil {
			return nil, errors.New("storage repeats a missing predecessor")
		}
		return previous, nil
	}
	s := &biomeStorage{bits: bits}
	words := wordCount(bits)
	s.words = in.Next(words * 4)
	if len(s.words) != words*4 {
		return nil, errors.New("short storage")
	}
	var count int32 = 1
	if bits != 0 {
		if err := protocol.Varint32(in, &count); err != nil || count <= 0 {
			return nil, errors.New("bad palette size")
		}
	}
	s.palette = make([]int32, count)
	for i := range s.palette {
		if err := protocol.Varint32(in, &s.palette[i]); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// column returns the 256 values of layer y, in the order of the pre-1.18
// column biomes: index z<<4|x.
func (s *biomeStorage) column(y int) []byte {
	out := make([]byte, 256)
	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			out[z<<4|x] = byte(s.at(x<<8 | z<<4 | y))
		}
	}
	return out
}

func (s *biomeStorage) at(index int) int32 {
	if s.bits == 0 {
		return s.palette[0]
	}
	perWord := 32 / int(s.bits)
	word := index / perWord
	off := word * 4
	v := uint32(s.words[off]) | uint32(s.words[off+1])<<8 | uint32(s.words[off+2])<<16 | uint32(s.words[off+3])<<24
	i := (v >> (uint(index%perWord) * uint(s.bits))) & (1<<s.bits - 1)
	if int(i) >= len(s.palette) {
		return 0
	}
	return s.palette[i]
}

func wordCount(bits byte) int {
	if bits == 0 {
		return 0
	}
	perWord := 32 / int(bits)
	return (4096 + perWord - 1) / perWord
}
