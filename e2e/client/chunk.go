package main

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// checkLegacyChunk parses a pre-1.18 LevelChunk payload the way those clients
// do: version 8 sub-chunks, 256 column biomes, the border block list and block
// entity NBT, and requires every byte to be used.
func checkLegacyChunk(count uint32, payload []byte) error {
	if count > 16 {
		return fmt.Errorf("%d sub-chunks, the world has 16", count)
	}
	in := bytes.NewBuffer(payload)
	for i := uint32(0); i < count; i++ {
		ver, _ := in.ReadByte()
		n, err := in.ReadByte()
		if err != nil || ver != 8 {
			return fmt.Errorf("sub-chunk %d: version %d", i, ver)
		}
		for j := byte(0); j < n; j++ {
			if err := skipStorage(in); err != nil {
				return fmt.Errorf("sub-chunk %d storage %d: %v", i, j, err)
			}
		}
	}
	if len(in.Next(256)) != 256 {
		return errors.New("column biomes missing")
	}
	border, err := in.ReadByte()
	if err != nil {
		return errors.New("border block count missing")
	}
	in.Next(int(border))
	dec := nbt.NewDecoderWithEncoding(in, nbt.NetworkLittleEndian)
	for in.Len() > 0 {
		var m map[string]interface{}
		if err := dec.Decode(&m); err != nil {
			return fmt.Errorf("block entity: %v", err)
		}
	}
	return nil
}

func skipStorage(in *bytes.Buffer) error {
	header, err := in.ReadByte()
	if err != nil {
		return err
	}
	bits := header >> 1
	if bits != 0 {
		perWord := 32 / int(bits)
		words := (4096 + perWord - 1) / perWord
		if len(in.Next(words*4)) != words*4 {
			return errors.New("short storage")
		}
	}
	var n int32 = 1
	if bits != 0 {
		if err := protocol.Varint32(in, &n); err != nil {
			return err
		}
	}
	for i := int32(0); i < n; i++ {
		var v int32
		if err := protocol.Varint32(in, &v); err != nil {
			return err
		}
	}
	return nil
}
