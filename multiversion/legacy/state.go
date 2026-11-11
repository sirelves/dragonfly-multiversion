package legacy

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// connState is what a delta remembers about one client between packets.
type connState struct {
	// stopFlying is set when a client before 1.19.0 reported it stopped flying,
	// and consumed by its next PlayerAuthInput.
	stopFlying atomic.Bool

	// sent holds the last packets sent to the client when MV_TRACE is set.
	mu   sync.Mutex
	sent []string
}

// trace is set by MV_TRACE=1: every legacy connection then keeps its last
// traceDepth outgoing packets and logs them when it closes, which is what a
// client crash leaves to look at.
var trace = os.Getenv("MV_TRACE") != ""

const traceDepth = 64

func (st *connState) record(pk packet.Packet) {
	if !trace {
		return
	}
	st.mu.Lock()
	defer st.mu.Unlock()
	if len(st.sent) == traceDepth {
		st.sent = st.sent[1:]
	}
	st.sent = append(st.sent, fmt.Sprintf("%T", pk))
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
			if trace {
				s := st.(*connState)
				s.mu.Lock()
				defer s.mu.Unlock()
				slog.Info("legacy connection closed; last packets sent", "name", conn.IdentityData().DisplayName,
					"version", conn.ClientData().GameVersion, "packets", strings.Join(s.sent, " "))
			}
		}()
	}
	return st.(*connState)
}
