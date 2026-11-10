package raknet

import "testing"

func TestAcceptedProtocols(t *testing.T) {
	for v, want := range map[byte]bool{9: false, 10: true, 11: true, 12: false} {
		if acceptedProtocol(v) != want {
			t.Errorf("protocol %d: accepted = %v", v, !want)
		}
	}
}
