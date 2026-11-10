package v786

import (
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// Before 1.16.220 (431) an item was written inline: varint network ID; unless
// air, varint aux (metadata << 8 | count, metadata 0x7fff for any), the user
// data (int16 -1, version 1 and network NBT, or int16 0), the can-place-on and
// can-break lists with varint counts, and for shields a varint64 blocking tick.
// The ID translation to the client registry happens around these calls.

// WriteItem407 writes x in the pre-431 item form.
func WriteItem407(w *protocol.Writer, shield int32, x *protocol.ItemStack) {
	w.Varint32(&x.NetworkID)
	if x.NetworkID == 0 {
		return
	}
	meta := x.MetadataValue
	if meta > 0x7fff {
		meta = 0x7fff
	}
	aux := int32(meta<<8) | int32(x.Count&0xff)
	w.Varint32(&aux)
	if len(x.NBTData) != 0 {
		marker, version := int16(-1), uint8(1)
		w.Int16(&marker)
		w.Uint8(&version)
		w.NBT(&x.NBTData, nbt.NetworkLittleEndian)
	} else {
		var marker int16
		w.Int16(&marker)
	}
	stringsVarint32(w, &x.CanBePlacedOn)
	stringsVarint32(w, &x.CanBreak)
	if x.NetworkID == shield {
		tick := x.BlockingTick
		w.Varint64(&tick)
	}
}

// ReadItem407 reads an item of the pre-431 form into x.
func ReadItem407(r *protocol.Reader, shield int32, x *protocol.ItemStack) {
	*x = protocol.ItemStack{NBTData: map[string]any{}}
	r.Varint32(&x.NetworkID)
	if x.NetworkID == 0 {
		return
	}
	var aux int32
	r.Varint32(&aux)
	x.MetadataValue, x.Count = uint32(aux>>8)&0x7fff, uint16(aux&0xff)
	var marker int16
	r.Int16(&marker)
	switch marker {
	case 0:
	case -1:
		var version uint8
		r.Uint8(&version)
		if version != 1 {
			r.UnknownEnumOption(version, "item user data version")
			return
		}
		r.NBT(&x.NBTData, nbt.NetworkLittleEndian)
	default:
		r.InvalidValue(marker, "item user data marker", "only -1 and 0 are known")
		return
	}
	stringsVarint32(r, &x.CanBePlacedOn)
	stringsVarint32(r, &x.CanBreak)
	if x.NetworkID == shield {
		r.Varint64(&x.BlockingTick)
	}
}

// stringsVarint32 reads or writes a string list with a varint32 count.
func stringsVarint32(io protocol.IO, x *[]string) {
	n := int32(len(*x))
	io.Varint32(&n)
	if n < 0 || n > 1024 {
		io.InvalidValue(n, "string list length", "must be between 0 and 1024")
		return
	}
	if len(*x) != int(n) {
		*x = make([]string, n)
	}
	for i := range *x {
		io.String(&(*x)[i])
	}
}
