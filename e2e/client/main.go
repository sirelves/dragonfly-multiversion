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
	"net"
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
	expect := flag.String("expect", "", "comma-separated packet types that must arrive, e.g. AdventureSettings")
	dump := flag.String("dump", "", "write every packet the server sends, as \"<id> <hex payload>\" lines, to this file")
	flag.Parse()

	sink := &errSink{}
	d := minecraft.Dialer{ErrorLog: newErrLog(sink)}
	d.IdentityData.DisplayName = *name
	// Real clients always send an identity UUID; the 1.16.0 tags do not make one
	// for offline logins.
	var h uint64 = 1469598103934665603
	for _, c := range *name {
		h = (h ^ uint64(c)) * 1099511628211
	}
	d.IdentityData.Identity = fmt.Sprintf("%08x-0000-4000-8000-%012x", uint32(h>>32), h&0xffffffffffff)
	if *dump != "" {
		f, err := os.Create(*dump)
		if err != nil {
			fmt.Println("FAIL dump:", err)
			os.Exit(1)
		}
		defer f.Close()
		var mu sync.Mutex
		// The server is whichever end the client never writes from: packets
		// read have the server as src, so it is the src that is not local.
		d.PacketFunc = func(hdr packet.Header, payload []byte, src, dst net.Addr) {
			if src.String() != *addr && !strings.HasSuffix(src.String(), (*addr)[strings.LastIndex(*addr, ":"):]) {
				return
			}
			mu.Lock()
			fmt.Fprintf(f, "%d %x\n", hdr.PacketID, payload)
			mu.Unlock()
		}
	}
	conn, err := dial(d, *addr, 20*time.Second)
	if err != nil {
		fail("dial: %v\n%s", err, strings.TrimSpace(sink.buf.String()))
	}
	defer conn.Close()
	if err := spawn(conn, 20*time.Second); err != nil {
		fail("spawn: %v\n%s", err, strings.TrimSpace(sink.buf.String()))
	}
	gd := conn.GameData()
	fmt.Printf("SPAWNED proto=%d pos=%v dim=%d\n", protocol.CurrentProtocol, gd.PlayerPosition, gd.Dimension)

	// Walk east for two seconds and say hello, so serverbound packets of this
	// version are decoded by the server too.
	go func() {
		pos := gd.PlayerPosition
		t := time.NewTicker(50 * time.Millisecond)
		defer t.Stop()
		for i := 0; i < 40; i++ {
			<-t.C
			pos[0] += 0.1
			if err := conn.WritePacket(&packet.PlayerAuthInput{Position: pos, Yaw: -90, HeadYaw: -90, InputMode: packet.InputModeMouse}); err != nil {
				return
			}
		}
		_ = conn.WritePacket(&packet.Text{TextType: packet.TextTypeChat, SourceName: *name, Message: "hello from " + *name})
	}()

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
		if lc, ok := pk.(*packet.LevelChunk); ok && protocol.CurrentProtocol < 475 && !lc.CacheEnabled {
			if err := checkLegacyChunk(lc.SubChunkCount, lc.RawPayload); err != nil {
				sink.Write([]byte("chunk: " + err.Error() + "\n"))
			}
		}
		if t, ok := pk.(*packet.Text); ok && strings.Contains(t.Message, "hello from") {
			fmt.Printf("CHAT %q\n", t.Message)
		}
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
	for _, want := range strings.Split(*expect, ",") {
		if want != "" && counts["*packet."+want] == 0 {
			sink.Write([]byte("expected packet never arrived: " + want + "\n"))
		}
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
