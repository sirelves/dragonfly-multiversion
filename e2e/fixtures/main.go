// Command fixtures encodes the conformance ItemStackRequest fixtures with the
// gophertunnel tag of one protocol. gen-fixtures.sh compiles it once per tag
// and merges the output into multiversion/conformance/testdata/item_stack.json.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func main() {
	sword := flag.Int("stone-sword", 0, "stone sword runtime ID of this protocol")
	rev := flag.String("revision", "", "gophertunnel commit")
	flag.Parse()

	slot := func(c byte, s byte, id int32) protocol.StackRequestSlotInfo {
		return protocol.StackRequestSlotInfo{ContainerID: c, Slot: s, StackNetworkID: id}
	}
	take := &protocol.TakeStackRequestAction{}
	take.Count, take.Source, take.Destination = 1, slot(60, 50, -3), slot(28, 0, 0)
	results := &protocol.CraftResultsDeprecatedStackRequestAction{TimesCrafted: 1,
		ResultItems: []protocol.ItemStack{{ItemType: protocol.ItemType{NetworkID: int32(*sword)}, Count: 1}}}
	actions := actionSet(take, results)

	out := map[string][]byte{}
	for name, as := range actions {
		pk := &packet.ItemStackRequest{Requests: []protocol.ItemStackRequest{{RequestID: -3, Actions: as, FilterCause: -1}}}
		buf := new(bytes.Buffer)
		pk.Marshal(protocol.NewWriter(buf, 0))
		out[name] = buf.Bytes()
	}
	damage := map[string][]byte{}
	for name, pk := range damagePackets() {
		buf := new(bytes.Buffer)
		pk.Marshal(protocol.NewWriter(buf, 0))
		damage[name] = buf.Bytes()
	}
	_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
		"item_stack": map[string]any{"revision": *rev, "stone_sword": *sword, "fixtures": out},
		"damage":     map[string]any{"revision": *rev, "packets": damage},
	})
}
