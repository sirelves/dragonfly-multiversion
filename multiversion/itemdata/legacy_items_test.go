package itemdata

import (
	"sort"
	"testing"

	_ "github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// TestLegacyItemCoverage checks that basic items survive the trip to every
// pre-1.20 client and back, and logs how many Dragonfly items each one lacks.
func TestLegacyItemCoverage(t *testing.T) {
	protos := make([]int, 0, len(legacyItemFiles))
	for p := range legacyItemFiles {
		protos = append(protos, int(p))
	}
	sort.Ints(protos)
	for _, p := range protos {
		proto := uint32(p)
		missing := 0
		for _, it := range world.Items() {
			name, meta := it.EncodeItem()
			id, ok := world.ItemRuntimeIDByName(name)
			if !ok || meta != 0 {
				continue
			}
			if old, _ := TranslateStack(proto, id, 0); old == 0 {
				missing++
			}
		}
		t.Logf("%d: %d Dragonfly items without a client equivalent", p, missing)
		for _, name := range []string{"minecraft:diamond_sword", "minecraft:stone", "minecraft:white_wool", "minecraft:oak_planks", "minecraft:apple", "minecraft:shield", "minecraft:spruce_log"} {
			id, ok := world.ItemRuntimeIDByName(name)
			if !ok {
				t.Fatalf("dragonfly has no %s", name)
			}
			old, meta := TranslateStack(proto, id, 0)
			if old == 0 {
				t.Errorf("%d cannot spell %s", p, name)
				continue
			}
			back, backMeta := ReverseTranslateStack(proto, old, meta)
			if back != id || backMeta != 0 {
				t.Errorf("%d: %s -> (%d,%d) -> (%d,%d), want (%d,0)", p, name, old, meta, back, backMeta, id)
			}
		}
	}
}
