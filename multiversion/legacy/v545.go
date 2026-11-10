package legacy

import (
	"fmt"

	"github.com/df-mc/dragonfly/multiversion/itemdata"
	"github.com/df-mc/dragonfly/multiversion/v630"
	"github.com/df-mc/dragonfly/multiversion/v786"
	"github.com/df-mc/dragonfly/server/item/recipe"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// 1.19.21 (545) and 1.19.20 (544) → 1.19.30 (554), CloudburstMC v554:
//   - Recipe ingredients were bare items (v786 writes that form below 554),
//     so ingredients naming an item tag are expanded here: one ingredient with
//     any metadata when the tag maps to a single legacy item, one recipe per
//     legacy item otherwise.
//   - ItemStackRequest lost FilterCause, StructureSettings lost
//     AllowNonTickingChunks and StructureBlockUpdate lost Waterlogged (gated
//     in v786).
//   - Text had no ObjectAnnouncement type yet.
//   - RakNet 10 and no RequestNetworkSettings: handled by the listener.
var delta545 = delta{
	proto: 545,
	fromLatest: func(proto uint32, pk packet.Packet) []packet.Packet {
		switch pk := pk.(type) {
		case *v630.CraftingData:
			return []packet.Packet{expandTagRecipes(proto, pk)}
		case *packet.Text:
			if pk.TextType == packet.TextTypeObjectAnnouncement {
				cp := *pk
				cp.TextType = packet.TextTypeObject
				return []packet.Packet{&cp}
			}
		}
		return keep(pk)
	},
	toLatest: func(_ uint32, pk packet.Packet) []packet.Packet {
		if pk, ok := pk.(*v786.ItemStackRequest); ok {
			for i := range pk.Requests {
				for _, a := range pk.Requests[i].Actions {
					stripRecipeVariant(a)
				}
			}
		}
		return keep(pk)
	},
}

// Recipe variants made from one tagged recipe share its network ID in the low
// bits; the variant index sits above variantShift.
const variantShift = 20

func stripRecipeVariant(a protocol.StackRequestAction) {
	const mask = 1<<variantShift - 1
	switch a := a.(type) {
	case *protocol.CraftRecipeStackRequestAction:
		a.RecipeNetworkID &= mask
	case *protocol.AutoCraftRecipeStackRequestAction:
		a.RecipeNetworkID &= mask
	case *protocol.CraftRecipeOptionalStackRequestAction:
		a.RecipeNetworkID &= mask
	}
}

func expandTagRecipes(proto uint32, cd *v630.CraftingData) *v630.CraftingData {
	out := *cd
	out.Recipes = make([]v630.Recipe, 0, len(cd.Recipes))
	for _, r := range cd.Recipes {
		var inputs *[]protocol.ItemDescriptorCount
		switch r.Kind {
		case 0, 5, 6:
			inputs = &r.Shapeless.Input
		case 1, 7:
			inputs = &r.Shaped.Input
		default:
			out.Recipes = append(out.Recipes, r)
			continue
		}
		variants, ok := tagVariants(proto, *inputs)
		if !ok {
			continue // a tag this client has no item for
		}
		for k, v := range variants {
			cp := r
			in := v
			switch r.Kind {
			case 0, 5, 6:
				cp.Shapeless.Input = in
				renameVariant(&cp.Shapeless.RecipeID, &cp.Shapeless.UUID, &cp.Shapeless.RecipeNetworkID, k)
			default:
				cp.Shaped.Input = in
				renameVariant(&cp.Shaped.RecipeID, &cp.Shaped.UUID, &cp.Shaped.RecipeNetworkID, k)
			}
			out.Recipes = append(out.Recipes, cp)
		}
	}
	return &out
}

func renameVariant(id *string, u *uuid.UUID, network *uint32, k int) {
	if k == 0 {
		return
	}
	*id = fmt.Sprintf("%s_v%d", *id, k)
	u[15] ^= byte(k)
	*network |= uint32(k) << variantShift
}

// tagVariants returns the ingredient lists to send instead of inputs, one per
// combination of legacy items the tags in inputs stand for. Every slot naming
// the same tag gets the same item, as in the old vanilla recipes.
func tagVariants(proto uint32, inputs []protocol.ItemDescriptorCount) ([][]protocol.ItemDescriptorCount, bool) {
	options := map[string][]protocol.DefaultItemDescriptor{}
	var order []string
	for _, in := range inputs {
		t, ok := in.Descriptor.(*protocol.ItemTagItemDescriptor)
		if !ok {
			continue
		}
		if _, seen := options[t.Tag]; seen {
			continue
		}
		opts := legacyTagItems(proto, t.Tag)
		if len(opts) == 0 {
			return nil, false
		}
		options[t.Tag] = opts
		order = append(order, t.Tag)
	}
	combos := [][]protocol.ItemDescriptorCount{inputs}
	for _, tag := range order {
		var next [][]protocol.ItemDescriptorCount
		for _, base := range combos {
			for _, d := range options[tag] {
				cp := make([]protocol.ItemDescriptorCount, len(base))
				for i, in := range base {
					cp[i] = in
					if t, ok := in.Descriptor.(*protocol.ItemTagItemDescriptor); ok && t.Tag == tag {
						d := d
						cp[i] = protocol.ItemDescriptorCount{Descriptor: &d, Count: in.Count}
					}
				}
				next = append(next, cp)
			}
		}
		combos = next
		if len(combos) > 64 {
			return combos[:64], true
		}
	}
	return combos, true
}

// legacyTagItems lists the items of tag as this client spells them: members
// that become the same legacy item collapse into one any-metadata descriptor.
func legacyTagItems(proto uint32, tag string) []protocol.DefaultItemDescriptor {
	metas := map[int32]map[uint32]bool{}
	names := map[int32]string{}
	var ids []int32
	for _, name := range recipe.ItemTagItems(tag) {
		native, ok := world.ItemRuntimeIDByName(name)
		if !ok {
			continue
		}
		id, meta := itemdata.TranslateStack(proto, native, 0)
		if id == 0 {
			continue
		}
		if metas[id] == nil {
			metas[id] = map[uint32]bool{}
			names[id] = name
			ids = append(ids, id)
		}
		metas[id][meta] = true
	}
	out := make([]protocol.DefaultItemDescriptor, 0, len(ids))
	for _, id := range ids {
		d := protocol.DefaultItemDescriptor{Name: names[id]}
		if len(metas[id]) > 1 {
			d.MetadataValue = 0x7fff
		}
		out = append(out, d)
	}
	return out
}
