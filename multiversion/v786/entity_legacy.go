package v786

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// Entity metadata keys and flag bits moved before 1.20.0: flags inserted at 46
// and 61 shifted the rest, keys were inserted at 60 and 120 and later removed
// again. Dragonfly writes gophertunnel's (current) numbering;
// legacyEntityMetadata rewrites it for a pre-589 protocol, dropping what that
// protocol has no slot for. Flags 0–63 live in key 0 and 64–127 in key 92.
func legacyEntityMetadata(proto uint32, x protocol.EntityMetadata) protocol.EntityMetadata {
	keys, flags := legacyEntityKeys[proto], legacyEntityFlags[proto]
	if keys == nil && flags == nil {
		return x
	}
	out := make(protocol.EntityMetadata, len(x))
	for k, v := range x {
		if k == protocol.EntityDataKeyFlags || k == protocol.EntityDataKeyFlagsTwo {
			continue
		}
		if to, ok := keys[k]; ok {
			if to < 0 {
				continue
			}
			k = uint32(to)
		}
		out[k] = v
	}
	var lo, hi int64
	set := func(i int16) {
		if i < 64 {
			lo |= 1 << uint(i)
		} else if i < 128 {
			hi |= 1 << uint(i-64)
		}
	}
	for key, base := range map[uint32]uint8{protocol.EntityDataKeyFlags: 0, protocol.EntityDataKeyFlagsTwo: 64} {
		bits, ok := x[key].(int64)
		if !ok {
			continue
		}
		for b := uint8(0); b < 64; b++ {
			if bits&(1<<b) == 0 {
				continue
			}
			i := base + b
			to, moved := flags[i]
			if !moved {
				set(int16(i))
			} else if to >= 0 {
				set(to)
			}
		}
	}
	if _, ok := x[protocol.EntityDataKeyFlags]; ok {
		out[protocol.EntityDataKeyFlags] = lo
	}
	if _, ok := x[protocol.EntityDataKeyFlagsTwo]; ok || hi != 0 {
		out[keyOf(protocol.EntityDataKeyFlagsTwo, keys)] = hi
	}
	return out
}

func keyOf(k uint32, keys map[uint32]int32) uint32 {
	if to, ok := keys[k]; ok && to >= 0 {
		return uint32(to)
	}
	return k
}
