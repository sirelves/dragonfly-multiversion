package v786

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// Stack request action type IDs before 1.18.10 (486), as old → 486 numbering,
// evaluated from the CloudburstMC ITEM_STACK_REQUEST_TYPES chains. 486 added
// the item container actions at 7 and 8 and shifted the rest up by two; the
// older eras also lacked some crafting actions. IDs 0–6 never moved.
var (
	stackActions471 = map[byte]byte{7: 9, 8: 10, 9: 11, 10: 12, 11: 13, 12: 14, 13: 15, 14: 16, 15: 17, 16: 18, 17: 19}
	stackActions428 = map[byte]byte{7: 9, 8: 10, 9: 11, 10: 12, 11: 13, 12: 14, 13: 15, 14: 18, 15: 19}
	stackActions422 = map[byte]byte{7: 9, 8: 10, 9: 12, 10: 13, 11: 14, 12: 15, 13: 18, 14: 19}
	stackActions407 = map[byte]byte{7: 9, 8: 10, 9: 12, 10: 13, 11: 14, 12: 18, 13: 19}
)

func stackActionTable(proto uint32) map[byte]byte {
	switch {
	case proto >= 471:
		return stackActions471
	case proto >= 428:
		return stackActions428
	case proto >= 422:
		return stackActions422
	default:
		return stackActions407
	}
}

// legacyStackActionID reads (reading) or writes the action type ID of a
// pre-486 protocol, translating between its numbering and the current one. It
// reports false for an ID that protocol has no action for.
func legacyStackActionID(r protocol.IO, proto uint32, id *byte, reading bool) bool {
	table := stackActionTable(proto)
	if reading {
		var raw byte
		r.Uint8(&raw)
		if raw <= 6 {
			*id = raw
			return true
		}
		v, ok := table[raw]
		*id = v
		return ok
	}
	if *id <= 6 {
		r.Uint8(id)
		return true
	}
	for old, cur := range table {
		if cur == *id {
			r.Uint8(&old)
			return true
		}
	}
	return false
}
