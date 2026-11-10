package legacy

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// 1.19.80 (582) → 1.20.0 (589), gophertunnel v1.29.0 → v1.30.0:
//   - StartGame gained ServerAuthoritativeSound (gated in v766).
//   - Emote gained XUID and PlatformID (gated in v712).
//   - UnlockedRecipes replaced the NewUnlocks bool with an UnlockType enum.
var delta582 = delta{
	proto: 582,
	fromLatest: func(_ uint32, pk packet.Packet) []packet.Packet {
		if pk, ok := pk.(*packet.UnlockedRecipes); ok {
			return []packet.Packet{&UnlockedRecipes582{NewUnlocks: pk.UnlockType == packet.UnlockedRecipesTypeNewlyUnlocked, Recipes: pk.Recipes}}
		}
		return keep(pk)
	},
}

// UnlockedRecipes582 is UnlockedRecipes before 1.20.0.
type UnlockedRecipes582 struct {
	NewUnlocks bool
	Recipes    []string
}

func (*UnlockedRecipes582) ID() uint32 { return packet.IDUnlockedRecipes }

func (pk *UnlockedRecipes582) Marshal(io protocol.IO) {
	io.Bool(&pk.NewUnlocks)
	protocol.FuncSlice(io, &pk.Recipes, io.String)
}
