// Command client is a headless Bedrock client used by the e2e suite. It is
// compiled once per protocol against the gophertunnel tag that shipped that
// protocol, so every byte it reads and writes uses the real wire format of that
// client version. It joins the server offline, spawns, walks a little and then
// reports every packet it could not decode.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type errSink struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (e *errSink) Write(p []byte) (int, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.buf.Write(p)
}

func main() {
	addr := flag.String("addr", "127.0.0.1:19132", "server address")
	name := flag.String("name", "e2e", "display name")
	dur := flag.Duration("for", 6*time.Second, "time to stay connected after spawn")
	flag.Parse()

	sink := &errSink{}
	d := minecraft.Dialer{ErrorLog: newErrLog(sink)}
	d.IdentityData.DisplayName = *name
	conn, err := d.DialTimeout("raknet", *addr, 20*time.Second)
	if err != nil {
		fail("dial: %v", err)
	}
	defer conn.Close()
	if err := conn.DoSpawnTimeout(20 * time.Second); err != nil {
		fail("spawn: %v", err)
	}
	gd := conn.GameData()
	fmt.Printf("SPAWNED proto=%d pos=%v dim=%d\n", protocol.CurrentProtocol, gd.PlayerPosition, gd.Dimension)

	counts := map[string]int{}
	deadline := time.Now().Add(*dur)
	_ = conn.SetReadDeadline(deadline)
	for time.Now().Before(deadline) {
		pk, err := conn.ReadPacket()
		if err != nil {
			if !strings.Contains(err.Error(), "deadline") && !strings.Contains(err.Error(), "timeout") {
				sink.Write([]byte("read: " + err.Error() + "\n"))
			}
			break
		}
		counts[fmt.Sprintf("%T", pk)]++
		if dc, ok := pk.(*packet.Disconnect); ok {
			sink.Write([]byte("disconnected: " + dc.Message + "\n"))
			break
		}
	}
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("PACKET %s %d\n", k, counts[k])
	}
	if s := strings.TrimSpace(sink.buf.String()); s != "" {
		fmt.Println("ERRORS")
		fmt.Println(s)
		os.Exit(2)
	}
	fmt.Println("OK")
}

func fail(format string, a ...any) {
	fmt.Printf("FAIL "+format+"\n", a...)
	os.Exit(1)
}
