package legacy

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// 1.18.30 (503) → 1.19.0 (527), CloudburstMC v527:
//   - PlayerAction gained ResultPosition (gated in v786), PlayerAuthInput the
//     interaction model (gated in v630), StartGame PropertyData and
//     WorldTemplateID (gated in v766).
//   - RequestAbility did not exist: the client reported flight through a
//     serverbound AdventureSettings. Taking off becomes RequestAbility; landing
//     sets StopFlying on its next PlayerAuthInput, the only stop signal
//     Dragonfly reads.
var delta503 = delta{
	proto: 503,
	packets: func(_ uint32, pool packet.Pool, listener bool) {
		if listener {
			pool[packet.IDAdventureSettings] = func() packet.Packet { return &packet.AdventureSettings{} }
		}
	},
	client: func(_ uint32, pk packet.Packet, st *connState) []packet.Packet {
		if a, ok := pk.(*packet.AdventureSettings); ok {
			if a.Flags&packet.AdventureFlagFlying != 0 {
				return []packet.Packet{&packet.RequestAbility{Ability: packet.AbilityFlying, Value: true}}
			}
			st.stopFlying.Store(true)
			return nil
		}
		return keep(pk)
	},
	latest: func(_ uint32, pk packet.Packet, st *connState) []packet.Packet {
		if in, ok := pk.(*packet.PlayerAuthInput); ok && st.stopFlying.Swap(false) {
			if in.InputData.Len() <= packet.InputFlagStopFlying {
				grown := protocol.NewInputFlags(packet.InputFlagCount)
				for i := 0; i < in.InputData.Len(); i++ {
					if in.InputData.Load(i) {
						grown.Set(i)
					}
				}
				in.InputData = grown
			}
			in.InputData.Set(packet.InputFlagStopFlying)
		}
		return keep(pk)
	},
}
