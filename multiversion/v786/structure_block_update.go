/*
 _____               _____
|_   _| __ _   _  __|_   _|__ _ __ ___  _ __ ___
  | || '__| | | |/ _ \| |/ _ \ '_ ` _ \| '_ ` _ \
  | || |  | |_| |  __/| |  __/ | | | | | | | | | |
  |_||_|   \__,_|\___||_|\___|_| |_| |_|_| |_| |_|

 _____ __  __       ____                               __ _
|_   _|  \/  |     |  _ \ _ __ __ _  __ _  ___  _ __  / _| |_   _
  | | | |\/| |_____| | | | '__/ _` |/ _` |/ _ \| '_ \| |_| | | | |
  | | | |  | |_____| |_| | | | (_| | (_| | (_) | | | |  _| | |_| |
  |_| |_|  |_|     |____/|_|  \__,_|\__, |\___/|_| |_|_| |_|\__, |
                                    |___/                   |___/

@author TrueTemm
@link   https://github.com/TrueTemm
TM-Dragonfly Project
*/

package v786

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	StructureBlockData = iota
	StructureBlockSave
	StructureBlockLoad
	StructureBlockCorner
	StructureBlockInvalid
	StructureBlockExport
)

const (
	StructureRedstoneSaveModeMemory = iota
	StructureRedstoneSaveModeDisk
)

type StructureBlockUpdate struct {
	Position protocol.BlockPos

	StructureName string

	FilteredStructureName string

	DataField string

	IncludePlayers bool

	ShowBoundingBox bool

	StructureBlockType int32

	Settings protocol.StructureSettings

	RedstoneSaveMode int32

	ShouldTrigger bool

	Waterlogged bool
}

func (*StructureBlockUpdate) ID() uint32 {
	return IDStructureBlockUpdate
}

func (pk *StructureBlockUpdate) Marshal(io protocol.IO) {
	UBlockPos786(io, &pk.Position)
	io.String(&pk.StructureName)
	io.String(&pk.FilteredStructureName)
	io.String(&pk.DataField)
	io.Bool(&pk.IncludePlayers)
	io.Bool(&pk.ShowBoundingBox)
	io.Varint32(&pk.StructureBlockType)
	StructureSettings786(io, &pk.Settings)
	io.Varint32(&pk.RedstoneSaveMode)
	io.Bool(&pk.ShouldTrigger)
	if p := ProtoOf(io); p == 0 || p >= 554 { // added in 554
		io.Bool(&pk.Waterlogged)
	}
}

// StructureSettings786 is protocol.StructureSettings with the fields that older
// protocols lack left out.
func StructureSettings786(io protocol.IO, x *protocol.StructureSettings) {
	io.String(&x.PaletteName)
	io.Bool(&x.IgnoreEntities)
	io.Bool(&x.IgnoreBlocks)
	if p := ProtoOf(io); p == 0 || p >= 554 { // added in 554
		io.Bool(&x.AllowNonTickingChunks)
	}
	io.BlockPos(&x.Size)
	io.BlockPos(&x.Offset)
	io.ActorUniqueID(&x.LastEditingPlayerUniqueID)
	io.Uint8(&x.Rotation)
	io.Uint8(&x.Mirror)
	if p := ProtoOf(io); p == 0 || p >= 440 { // added in 440
		io.Uint8(&x.AnimationMode)
		io.Float32(&x.AnimationDuration)
	}
	io.Float32(&x.Integrity)
	io.Uint32(&x.Seed)
	io.Vec3(&x.Pivot)
}
