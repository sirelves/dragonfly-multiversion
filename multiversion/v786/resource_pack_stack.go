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

type ResourcePackStack struct {
	TexturePackRequired bool

	BehaviourPacks []protocol.StackResourcePack

	TexturePacks []protocol.StackResourcePack

	BaseGameVersion string

	Experiments []protocol.ExperimentData

	ExperimentsPreviouslyToggled bool

	IncludeEditorPacks bool
}

func (*ResourcePackStack) ID() uint32 {
	return IDResourcePackStack
}

func (pk *ResourcePackStack) Marshal(io protocol.IO) {
	io.Bool(&pk.TexturePackRequired)
	protocol.Slice(io, &pk.BehaviourPacks)
	protocol.Slice(io, &pk.TexturePacks)
	if p := ProtoOf(io); p != 0 && p < 419 { // an experimental flag before 419
		experimental := len(pk.Experiments) != 0
		io.Bool(&experimental)
		io.String(&pk.BaseGameVersion)
		return
	}
	io.String(&pk.BaseGameVersion)
	protocol.SliceUint32Length(io, &pk.Experiments)
	io.Bool(&pk.ExperimentsPreviouslyToggled)
	if p := ProtoOf(io); p == 0 || p >= 671 { // added in 671
		io.Bool(&pk.IncludeEditorPacks)
	}
}
