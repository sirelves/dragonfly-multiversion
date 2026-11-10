package itemdata

import (
	"embed"
	"fmt"
	"sync"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// Item runtime IDs of the 1.16.0 to 1.19.80 clients. 1.16.0 and 1.16.20 have
// no runtime registry: their IDs are the legacy numeric item IDs, which the
// server announces in StartGame all the same.
//
//go:embed runtime_item_states_407.json runtime_item_states_408.json runtime_item_states_419.json
//go:embed runtime_item_states_422.json runtime_item_states_428.json runtime_item_states_431.json
//go:embed runtime_item_states_440.json runtime_item_states_448.json runtime_item_states_465.json
//go:embed runtime_item_states_471.json runtime_item_states_475.json runtime_item_states_486.json
//go:embed runtime_item_states_503.json runtime_item_states_527.json runtime_item_states_534.json
//go:embed runtime_item_states_544.json runtime_item_states_545.json runtime_item_states_554.json
//go:embed runtime_item_states_557.json runtime_item_states_560.json runtime_item_states_567.json
//go:embed runtime_item_states_575.json runtime_item_states_582.json
var legacyStates embed.FS

// legacyItemFiles maps each pre-1.20 protocol to the file holding its items.
var legacyItemFiles = map[uint32]uint32{
	407: 407, 408: 408, 419: 419, 422: 422, 428: 428, 431: 431, 440: 440, 448: 448,
	465: 465, 471: 471, 475: 475, 486: 486, 503: 503, 527: 527, 534: 534, 544: 544,
	545: 545, 554: 554, 557: 557, 560: 560, 567: 567, 568: 567, 575: 575, 582: 582,
}

func init() {
	for proto, file := range legacyItemFiles {
		var (
			once    sync.Once
			entries []protocol.ItemEntry
		)
		load := func() []protocol.ItemEntry {
			once.Do(func() {
				data, err := legacyStates.ReadFile(fmt.Sprintf("runtime_item_states_%d.json", file))
				if err != nil {
					panic("itemdata: " + err.Error())
				}
				entries = loadStatesOnly(data)
			})
			return entries
		}
		itemPalettes[proto] = &itemPalette{entries: load}
		legacyPalettes[proto] = &legacyPalette{}
	}
}
