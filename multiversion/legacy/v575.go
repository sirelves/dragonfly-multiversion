package legacy

import (
	"github.com/df-mc/dragonfly/multiversion/v630"
	"github.com/df-mc/dragonfly/multiversion/v786"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// 1.19.70 (575) → 1.19.80 (582), gophertunnel v1.28.2 → v1.29.0:
//   - RequestChunkRadius gained MaxChunkRadius (gated in v786).
//   - StartGame gained CreatedInEditor, ExportedFromEditor and
//     UseBlockNetworkIDHashes (gated in v766).
//   - Command parameter type IDs moved (remapped for every protocol up to 575,
//     see commands.go).
//   - SmithingTransform recipes gained a template and SmithingTrim recipes
//     (kind 9) appeared. The server only accepts smithing with a template, so
//     older clients get no smithing recipes at all.
var delta575 = delta{
	proto: 575,
	fromLatest: func(proto uint32, pk packet.Packet) []packet.Packet {
		if ac, ok := pk.(*v786.AvailableCommands); ok {
			return []packet.Packet{remapCommandArgs(int32(proto), ac)}
		}
		if cd, ok := pk.(*v630.CraftingData); ok {
			out := *cd
			out.Recipes = make([]v630.Recipe, 0, len(cd.Recipes))
			for _, r := range cd.Recipes {
				if r.Kind != recipeSmithingTransform && r.Kind != recipeSmithingTrim {
					out.Recipes = append(out.Recipes, r)
				}
			}
			return []packet.Packet{&out}
		}
		return keep(pk)
	},
}

const (
	recipeSmithingTransform = 8
	recipeSmithingTrim      = 9
)
