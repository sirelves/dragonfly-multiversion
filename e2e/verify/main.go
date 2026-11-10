// Command verify cross-checks the server's encoders against the gophertunnel
// tag that shipped a protocol. It is compiled per tag by verify.sh.
//
//	verify -send send_<p>.txt   decodes every packet the server encoded for
//	                            this protocol, requires that no byte is left
//	                            over and that re-encoding gives the same bytes
//	verify -recv recv_<p>.txt   writes every packet of this version, filled
//	                            with non-zero values, for the server to decode
package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unsafe"
)

func main() {
	send := flag.String("send", "", "file written by TestBlindVerifySend")
	recv := flag.String("recv", "", "file to write for TestBlindVerifyRecv")
	ids := flag.Bool("ids", false, "print every packet ID this version knows")
	flag.Parse()
	if *ids {
		all := allIDs()
		sort.Slice(all, func(i, j int) bool { return all[i] < all[j] })
		for _, id := range all {
			fmt.Println(id)
		}
		return
	}
	if *recv != "" {
		writeRecv(*recv)
	}
	if *send != "" {
		os.Exit(checkSend(*send))
	}
}

func checkSend(path string) int {
	f, err := os.Open(path)
	if err != nil {
		fmt.Println("open:", err)
		return 1
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<20), 1<<26)
	bad, good := 0, 0
	for sc.Scan() {
		p := strings.Split(sc.Text(), "\t")
		if len(p) < 4 || p[2] != "ok" {
			if len(p) >= 3 && p[2] != "ok" {
				fmt.Printf("BAD %-36s server: %s\n", p[0], p[2])
				bad++
			}
			continue
		}
		id, _ := strconv.Atoi(p[1])
		raw, _ := hex.DecodeString(p[3])
		if msg := check(uint32(id), raw); msg != "" {
			fmt.Printf("BAD %-36s id=%d %s\n", p[0], id, msg)
			bad++
			continue
		}
		good++
	}
	fmt.Printf("SEND ok=%d bad=%d\n", good, bad)
	if bad > 0 {
		return 2
	}
	return 0
}

// check decodes raw as packet id and re-encodes it.
func check(id uint32, raw []byte) (msg string) {
	defer func() {
		if e := recover(); e != nil {
			msg = fmt.Sprintf("decode panic: %v", e)
		}
	}()
	rest, again, err := roundTrip(id, raw)
	switch {
	case err != nil:
		return err.Error()
	case rest != 0:
		return fmt.Sprintf("%d of %d bytes unread", rest, len(raw))
	case !bytes.Equal(again, raw):
		return fmt.Sprintf("re-encode differs:\n      server %x\n      client %x", raw, again)
	}
	return ""
}

func writeRecv(path string) {
	filled = os.Getenv("BLINDVERIFY_FILL") != ""
	var lines []string
	for _, id := range poolIDs() {
		name, raw, err := encodeFilled(id)
		status := "ok"
		if err != nil {
			status = "ENCODE_PANIC:" + strings.ReplaceAll(err.Error(), "\t", " ")
		}
		lines = append(lines, fmt.Sprintf("%d\t%s\t%s\t%s", id, name, status, hex.EncodeToString(raw)))
	}
	sort.Strings(lines)
	_ = os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}

var filled bool

// fill sets every reachable field to a small non-zero value.
func fill(v reflect.Value, depth int) {
	if depth > 7 || !v.IsValid() {
		return
	}
	if !v.CanSet() {
		if !v.CanAddr() {
			return
		}
		v = reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
	}
	switch v.Kind() {
	case reflect.Bool:
		v.SetBool(true)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(1)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v.SetUint(1)
	case reflect.Float32, reflect.Float64:
		v.SetFloat(1.5)
	case reflect.String:
		v.SetString("a")
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			v.SetBytes([]byte{1, 2})
			return
		}
		s := reflect.MakeSlice(v.Type(), 1, 1)
		fill(s.Index(0), depth+1)
		v.Set(s)
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			fill(v.Index(i), depth+1)
		}
	case reflect.Map:
		m := reflect.MakeMap(v.Type())
		k := reflect.New(v.Type().Key()).Elem()
		fill(k, depth+1)
		e := reflect.New(v.Type().Elem()).Elem()
		if e.Kind() == reflect.Interface {
			e.Set(reflect.ValueOf(int32(1)))
		} else {
			fill(e, depth+1)
		}
		m.SetMapIndex(k, e)
		v.Set(m)
	case reflect.Ptr:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		fill(v.Elem(), depth+1)
	case reflect.Struct:
		if v.Type().String() == "big.Int" || strings.HasSuffix(v.Type().Name(), "Bitset") {
			return
		}
		for i := 0; i < v.NumField(); i++ {
			fill(v.Field(i), depth+1)
		}
	}
}
