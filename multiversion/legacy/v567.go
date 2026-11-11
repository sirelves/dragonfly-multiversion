package legacy

import (
	"strconv"
	"strings"

	"github.com/df-mc/dragonfly/multiversion/v630"
	"github.com/df-mc/dragonfly/multiversion/v786"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// 1.19.60 (567) and 1.19.63 (568) → 1.19.70 (575).
var delta567 = delta{proto: 568}

// Two clients speak protocol 567: 1.19.60 and its hotfix 1.19.62, which added
// a flag at the end of every skin (OverrideAppearance) without a protocol
// bump. Only the game version in the login tells them apart, so this runs per
// connection: 1.19.60 clients get their skins without the flag. What they send
// (PlayerSkin, rare) is still read in the 1.19.62 form.
func skinsFor11960(conn *minecraft.Conn, pks []packet.Packet) {
	if conn == nil || !before11962(conn.ClientData().GameVersion) {
		return
	}
	for _, pk := range pks {
		var list *v786.PlayerList
		var skin *v786.PlayerSkin
		switch pk := pk.(type) { // v630 wraps the v786 packets
		case *v786.PlayerList:
			list = pk
		case *v630.PlayerList:
			list = &pk.PlayerList
		case *v786.PlayerSkin:
			skin = pk
		}
		if list != nil {
			for i := range list.Entries {
				list.Entries[i].Skin.NoOverrideAppearance = true
			}
		}
		if skin != nil {
			skin.Skin.NoOverrideAppearance = true
		}
	}
}

// before11962 reports whether a 1.19.x game version is older than 1.19.62.
func before11962(v string) bool {
	parts := strings.Split(v, ".")
	if len(parts) < 3 || parts[0] != "1" || parts[1] != "19" {
		return false
	}
	patch, err := strconv.Atoi(parts[2])
	return err == nil && patch < 62
}
