package legacy

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// 1.16.200 (422) → 1.16.210 (428), CloudburstMC v428:
//   - StartGame carried only the movement mode instead of the movement
//     settings, PlayerAuthInput had no interaction sections and skins no
//     PlayFab ID (gated in v630, v766 and v786).
//   - CameraShake had no action.
var delta422 = delta{
	proto: 422,
	fromLatest: func(_ uint32, pk packet.Packet) []packet.Packet {
		if pk, ok := pk.(*packet.CameraShake); ok {
			if pk.Action == packet.CameraShakeActionStop {
				return nil // stopping a shake did not exist yet
			}
			return []packet.Packet{&CameraShake422{Intensity: pk.Intensity, Duration: pk.Duration, Type: pk.Type}}
		}
		return keep(pk)
	},
}

// CameraShake422 is CameraShake before 1.16.210.
type CameraShake422 struct {
	Intensity, Duration float32
	Type                uint8
}

func (*CameraShake422) ID() uint32 { return packet.IDCameraShake }

func (pk *CameraShake422) Marshal(io protocol.IO) {
	io.Float32(&pk.Intensity)
	io.Float32(&pk.Duration)
	io.Uint8(&pk.Type)
}
