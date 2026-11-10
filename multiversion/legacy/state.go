package legacy

import (
	"sync"
	"sync/atomic"

	"github.com/sandertv/gophertunnel/minecraft"
)

// connState is what a delta remembers about one client between packets.
type connState struct {
	// stopFlying is set when a client before 1.19.0 reported it stopped flying,
	// and consumed by its next PlayerAuthInput.
	stopFlying atomic.Bool
}

var states sync.Map // *minecraft.Conn → *connState

// stateOf returns the state of conn, created on first use and dropped when the
// connection closes. A nil conn (tests, conversions outside a session) gets a
// fresh state every time.
func stateOf(conn *minecraft.Conn) *connState {
	if conn == nil {
		return &connState{}
	}
	if st, ok := states.Load(conn); ok {
		return st.(*connState)
	}
	st, loaded := states.LoadOrStore(conn, &connState{})
	if !loaded {
		go func() {
			<-conn.Context().Done()
			states.Delete(conn)
		}()
	}
	return st.(*connState)
}
