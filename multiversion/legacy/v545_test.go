package legacy

import (
	"testing"

	"github.com/df-mc/dragonfly/multiversion/itemdata"
	"github.com/df-mc/dragonfly/multiversion/v630"
	"github.com/df-mc/dragonfly/multiversion/v786"
	_ "github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func tagged(tag string, n int) []protocol.ItemDescriptorCount {
	in := make([]protocol.ItemDescriptorCount, n)
	for i := range in {
		in[i] = protocol.ItemDescriptorCount{Descriptor: &protocol.ItemTagItemDescriptor{Tag: tag}, Count: 1}
	}
	return in
}

func TestTagRecipesBecomeLegacyItems(t *testing.T) {
	planks := v630.Recipe{Kind: 1, Shaped: protocol.ShapedRecipe{RecipeID: "crafting_table", Width: 2, Height: 2, Input: tagged("minecraft:planks", 4), RecipeNetworkID: 7}}
	logs := v630.Recipe{Kind: 0, Shapeless: protocol.ShapelessRecipe{RecipeID: "planks", Input: tagged("minecraft:logs", 1), RecipeNetworkID: 9}}
	out := expandTagRecipes(545, &v630.CraftingData{Recipes: []v630.Recipe{planks, logs}})

	var table, fromLogs []v630.Recipe
	for _, r := range out.Recipes {
		switch r.Kind {
		case 1:
			table = append(table, r)
		case 0:
			fromLogs = append(fromLogs, r)
		}
	}
	// planks (meta 0-5) plus the separate crimson, warped and mangrove planks.
	if len(table) != 4 {
		t.Fatalf("crafting table: %d variants, want 4", len(table))
	}
	planksID, _ := itemdata.TranslateStack(545, mustID(t, "minecraft:oak_planks"), 0)
	wildcard := 0
	ids := map[int32]bool{}
	for _, r := range table {
		first := r.Shaped.Input[0].Descriptor.(*protocol.DefaultItemDescriptor)
		for _, in := range r.Shaped.Input {
			if d := in.Descriptor.(*protocol.DefaultItemDescriptor); *d != *first {
				t.Fatalf("variant mixes %v and %v", d, first)
			}
		}
		id, _ := itemdata.TranslateStack(545, mustID(t, first.Name), 0)
		ids[id] = true
		if id == planksID {
			if first.MetadataValue != 0x7fff {
				t.Fatalf("legacy planks ingredient has metadata %d, want any", first.MetadataValue)
			}
			wildcard++
		}
	}
	if wildcard != 1 || len(ids) != 4 {
		t.Fatalf("variants cover %d legacy items with %d planks wildcards", len(ids), wildcard)
	}
	if len(fromLogs) < 2 {
		t.Fatalf("logs: %d variants, want one per legacy log item", len(fromLogs))
	}
	seen := map[uint32]bool{}
	for _, r := range fromLogs {
		if seen[r.Shapeless.RecipeNetworkID] {
			t.Fatalf("duplicate network ID %d", r.Shapeless.RecipeNetworkID)
		}
		seen[r.Shapeless.RecipeNetworkID] = true
		a := &protocol.CraftRecipeStackRequestAction{RecipeNetworkID: r.Shapeless.RecipeNetworkID}
		req := &v786.ItemStackRequest{ItemStackRequest: packet.ItemStackRequest{Requests: []protocol.ItemStackRequest{{Actions: []protocol.StackRequestAction{a}}}}}
		delta545.toLatest(545, req)
		if a.RecipeNetworkID != 9 {
			t.Fatalf("variant network ID %d maps back to %d, want 9", r.Shapeless.RecipeNetworkID, a.RecipeNetworkID)
		}
	}
}

func mustID(t *testing.T, name string) int32 {
	id, ok := world.ItemRuntimeIDByName(name)
	if !ok {
		t.Fatalf("no item %s", name)
	}
	return id
}
