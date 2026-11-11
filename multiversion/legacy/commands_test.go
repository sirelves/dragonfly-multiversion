package legacy

import (
	"testing"

	"github.com/df-mc/dragonfly/multiversion/v786"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestCommandArgsReachOldNumbering(t *testing.T) {
	pk := &packet.AvailableCommands{Commands: []protocol.Command{{Name: "tp", Overloads: []protocol.CommandOverload{{Parameters: []protocol.CommandParameter{
		{Name: "x", Type: protocol.CommandArgValid | protocol.CommandArgTypeFloat},
		{Name: "who", Type: protocol.CommandArgValid | protocol.CommandArgTypeTarget},
		{Name: "pos", Type: protocol.CommandArgValid | protocol.CommandArgTypePosition},
		{Name: "mode", Type: protocol.CommandArgValid | protocol.CommandArgEnum | 3},
	}}}}}}
	// gophertunnel v1.10.0 (1.16.100): Float 2, Target 6, Position 37.
	ac := New(419).ConvertFromLatest(pk, nil)[0].(*v786.AvailableCommands)
	got := ac.Commands[0].Overloads[0].Parameters
	for i, want := range []uint32{protocol.CommandArgValid | 2, protocol.CommandArgValid | 6, protocol.CommandArgValid | 37, protocol.CommandArgValid | protocol.CommandArgEnum | 3} {
		if got[i].Type != want {
			t.Errorf("%s: type %#x, want %#x", got[i].Name, got[i].Type, want)
		}
	}
}
