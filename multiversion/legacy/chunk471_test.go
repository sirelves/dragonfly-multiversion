package legacy

import (
	"bytes"
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

func TestLegacyChunkKeepsOnlyZeroTo255(t *testing.T) {
	world.DefaultBlockRegistry.Finalize()
	reg := world.DefaultBlockRegistry
	stone, dirt, gold := world.BlockRuntimeID(block.Stone{}), world.BlockRuntimeID(block.Dirt{}), world.BlockRuntimeID(block.Gold{})
	c := chunk.New(reg, cube.Range{-64, 319})
	c.SetBlock(1, -30, 2, 0, dirt)  // below the old world: dropped
	c.SetBlock(1, 10, 2, 0, stone)  // sub-chunk 0
	c.SetBlock(3, 70, 4, 0, gold)   // sub-chunk 4
	c.SetBlock(3, 300, 4, 0, stone) // above the old world: dropped
	for x := uint8(0); x < 16; x++ {
		for z := uint8(0); z < 16; z++ {
			for y := int16(-64); y < 320; y++ {
				c.SetBiome(x, y, z, 1)
			}
			c.SetBiome(x, 64, z, uint32(x)+2) // the layer the column biome comes from
		}
	}
	data := chunk.Encode(c, chunk.NetworkEncoding)
	payload := new(bytes.Buffer)
	for _, s := range data.SubChunks {
		payload.Write(s)
	}
	payload.Write(data.Biomes)
	payload.WriteByte(0)
	enc := nbt.NewEncoderWithEncoding(payload, nbt.NetworkLittleEndian)
	_ = enc.Encode(map[string]any{"id": "Chest", "x": int32(1), "y": int32(10), "z": int32(2)})
	_ = enc.Encode(map[string]any{"id": "Chest", "x": int32(1), "y": int32(-30), "z": int32(2)})

	out, count, err := legacyChunk(payload.Bytes(), uint32(len(data.SubChunks)), 0)
	if err != nil {
		t.Fatal(err)
	}
	if count != 5 {
		t.Fatalf("%d sub-chunks, want 5 (y 0..79)", count)
	}
	in := bytes.NewBuffer(out)
	var palettes [][]int32
	for i := uint32(0); i < count; i++ {
		ver, _ := in.ReadByte()
		n, _ := in.ReadByte()
		if ver != 8 {
			t.Fatalf("sub-chunk %d has version %d", i, ver)
		}
		var ids []int32
		for j := byte(0); j < n; j++ {
			s, err := readStorage(in, nil)
			if err != nil {
				t.Fatalf("sub-chunk %d: %v", i, err)
			}
			ids = append(ids, s.palette...)
		}
		palettes = append(palettes, ids)
	}
	has := func(ids []int32, rid uint32) bool {
		for _, id := range ids {
			if uint32(id) == rid {
				return true
			}
		}
		return false
	}
	if !has(palettes[0], stone) || has(palettes[0], dirt) || !has(palettes[4], gold) {
		t.Fatalf("blocks landed in the wrong sub-chunks: %v", palettes)
	}
	biomes := in.Next(256)
	for x := 0; x < 16; x++ {
		if got := biomes[3<<4|x]; got != byte(x+2) {
			t.Fatalf("biome of column x=%d is %d, want %d", x, got, x+2)
		}
	}
	if b, _ := in.ReadByte(); b != 0 {
		t.Fatal("border block count missing")
	}
	var be map[string]any
	if err := nbt.NewDecoderWithEncoding(in, nbt.NetworkLittleEndian).Decode(&be); err != nil || be["y"] != int32(10) {
		t.Fatalf("block entity %v (%v)", be, err)
	}
	if in.Len() != 0 {
		t.Fatalf("%d bytes left: the block entity below 0 was kept", in.Len())
	}
}
