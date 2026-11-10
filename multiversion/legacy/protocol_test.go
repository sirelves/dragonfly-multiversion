package legacy

import (
	"github.com/df-mc/dragonfly/multiversion/v630"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestEveryProtocolHasKnownIDs(t *testing.T) {
	for _, id := range order {
		if len(knownIDs[id]) < 100 {
			t.Errorf("protocol %d: %d known packet IDs", id, len(knownIDs[id]))
		}
	}
}

func TestUnknownPacketsAreDropped(t *testing.T) {
	p := New(582)
	if out := p.ConvertFromLatest(&packet.LocatorBar{}, nil); len(out) != 0 {
		t.Fatalf("1.19.80 got %T, a packet from 1.21.80", out[0])
	}
	if out := p.ConvertFromLatest(&packet.SetTime{Time: 1}, nil); len(out) != 1 {
		t.Fatal("SetTime dropped")
	}
}

func TestSmithingRecipesNeedATemplate(t *testing.T) {
	in := &v630.CraftingData{Recipes: []v630.Recipe{{Kind: 0}, {Kind: recipeSmithingTransform}, {Kind: recipeSmithingTrim}}}
	for proto, want := range map[uint32]int{582: 3, 575: 1} {
		out := delta575.fromLatest(proto, in)[0].(*v630.CraftingData)
		if proto == 582 {
			out = in // 582 does not run delta575
		}
		if len(out.Recipes) != want {
			t.Errorf("%d: %d recipes, want %d", proto, len(out.Recipes), want)
		}
	}
	if len(in.Recipes) != 3 {
		t.Fatal("delta575 modified its input")
	}
}
