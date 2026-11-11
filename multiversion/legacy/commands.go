package legacy

import (
	"github.com/df-mc/dragonfly/multiversion/v786"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// remapCommandArgs rewrites the basic parameter types of an AvailableCommands
// from the 589 numbering into proto's (commandArgTypes, generated from each
// gophertunnel tag). An ID the client does not know can make it drop out when
// it builds its command list, so this is not cosmetic. Enum, soft enum and
// suffixed parameters keep their indices.
func remapCommandArgs(proto int32, pk *v786.AvailableCommands) *v786.AvailableCommands {
	table := commandArgTypes[proto]
	if len(table) == 0 {
		return pk
	}
	const special = protocol.CommandArgEnum | protocol.CommandArgSuffixed | protocol.CommandArgSoftEnum
	out := *pk
	out.Commands = make([]v786.Command786, len(pk.Commands))
	for i, c := range pk.Commands {
		c.Overloads = append([]protocol.CommandOverload(nil), c.Overloads...)
		for j, o := range c.Overloads {
			o.Parameters = append([]protocol.CommandParameter(nil), o.Parameters...)
			for k, p := range o.Parameters {
				if p.Type&special != 0 || p.Type&protocol.CommandArgValid == 0 {
					continue
				}
				if to, ok := table[p.Type&0xffff]; ok {
					o.Parameters[k].Type = p.Type&^0xffff | to
				}
			}
			c.Overloads[j] = o
		}
		out.Commands[i] = c
	}
	return &out
}
