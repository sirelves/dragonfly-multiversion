// Command mvgen converts a per-protocol BedrockData dump (one directory per
// protocol, e.g. IdotClub/BedrockData/v582) into the embedded data files the
// multiversion packages read:
//
//	multiversion/blockpalette/data/states_<p>.txt.gz   block states in runtime ID order
//	multiversion/itemdata/runtime_item_states_<p>.json item runtime IDs
//	multiversion/v630/data/<p>/*                        biomes, entities, sounds, particles
//
// Usage: go run ./cmd/mvgen -src <BedrockData dir> -proto 582 [-proto 575 ...]
package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

type protos []string

func (p *protos) String() string     { return strings.Join(*p, ",") }
func (p *protos) Set(v string) error { *p = append(*p, v); return nil }

func main() {
	src := flag.String("src", "", "BedrockData directory with v<protocol> subdirectories")
	root := flag.String("root", "multiversion", "multiversion package directory")
	var list protos
	flag.Var(&list, "proto", "protocol to convert (repeatable)")
	flag.Parse()
	if *src == "" || len(list) == 0 {
		flag.Usage()
		os.Exit(2)
	}
	for _, p := range list {
		if err := convert(filepath.Join(*src, "v"+p), *root, p); err != nil {
			fmt.Fprintf(os.Stderr, "mvgen %s: %v\n", p, err)
			os.Exit(1)
		}
	}
}

func convert(dir, root, proto string) error {
	if err := states(dir, root, proto); err != nil {
		return fmt.Errorf("block states: %w", err)
	}
	if err := items(dir, root, proto); err != nil {
		return fmt.Errorf("items: %w", err)
	}
	out := filepath.Join(root, "v630", "data", proto)
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	for _, f := range []string{"biome_definitions.nbt", "entity_identifiers.nbt", "level_sound_id_map.json", "particle_id_map.json"} {
		data, err := os.ReadFile(filepath.Join(dir, f))
		if errors.Is(err, os.ErrNotExist) {
			fmt.Printf("%s: no %s\n", proto, f)
			continue
		} else if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(out, f), data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// states writes canonical_block_states.nbt as one line per state, keeping the
// file order: that order is the runtime ID order of the client.
func states(dir, root, proto string) error {
	data, err := os.ReadFile(filepath.Join(dir, "canonical_block_states.nbt"))
	if err != nil {
		return err
	}
	list, err := decodeStates(data)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	for _, s := range list {
		buf.WriteString(s.Name)
		keys := make([]string, 0, len(s.States))
		for k := range s.States {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			buf.WriteByte('\t')
			buf.WriteString(k)
			buf.WriteByte('=')
			switch v := s.States[k].(type) {
			case uint8:
				buf.WriteString("b" + strconv.Itoa(int(v)))
			case int32:
				buf.WriteString("i" + strconv.Itoa(int(v)))
			case string:
				buf.WriteString("s" + v)
			default:
				return fmt.Errorf("%s: property %s has type %T", s.Name, k, v)
			}
		}
		buf.WriteByte('\n')
	}
	n := len(list)
	f, err := os.Create(filepath.Join(root, "blockpalette", "data", "states_"+proto+".txt.gz"))
	if err != nil {
		return err
	}
	defer f.Close()
	gz, _ := gzip.NewWriterLevel(f, gzip.BestCompression)
	if _, err := gz.Write(buf.Bytes()); err != nil {
		return err
	}
	fmt.Printf("%s: %d block states\n", proto, n)
	return gz.Close()
}

type itemState struct {
	Name           string `json:"name"`
	ID             int32  `json:"id"`
	ComponentBased bool   `json:"componentBased"`
}

// items writes the item runtime IDs. From 1.16.100 they come from
// required_item_list.json; before that the client used the legacy numeric IDs
// of item_id_map.json.
func items(dir, root, proto string) error {
	var list []itemState
	if data, err := os.ReadFile(filepath.Join(dir, "required_item_list.json")); err == nil {
		var m map[string]struct {
			RuntimeID      int32 `json:"runtime_id"`
			ComponentBased bool  `json:"component_based"`
		}
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}
		for name, e := range m {
			list = append(list, itemState{Name: name, ID: e.RuntimeID, ComponentBased: e.ComponentBased})
		}
	} else if data, err := os.ReadFile(filepath.Join(dir, "item_id_map.json")); err == nil {
		var m map[string]int32
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}
		for name, id := range m {
			list = append(list, itemState{Name: name, ID: id})
		}
	} else {
		return err
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
	out, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("%s: %d items\n", proto, len(list))
	return os.WriteFile(filepath.Join(root, "itemdata", "runtime_item_states_"+proto+".json"), append(out, '\n'), 0o644)
}

type blockState struct {
	Name    string         `nbt:"name"`
	States  map[string]any `nbt:"states"`
	Version int32          `nbt:"version"`
}

// decodeStates reads canonical_block_states.nbt. Since 1.16.100 it is a run of
// root compounds; the 1.16.0/1.16.20 dumps are one list of {block, id} entries.
func decodeStates(data []byte) ([]blockState, error) {
	r := bytes.NewReader(data)
	dec := nbt.NewDecoderWithEncoding(r, nbt.NetworkLittleEndian)
	if len(data) > 0 && data[0] == 9 { // TAG_List
		var legacy []struct {
			Block blockState `nbt:"block"`
			ID    int16      `nbt:"id"`
		}
		if err := dec.Decode(&legacy); err != nil {
			return nil, err
		}
		out := make([]blockState, len(legacy))
		for i, e := range legacy {
			out[i] = e.Block
		}
		return out, nil
	}
	var out []blockState
	for r.Len() > 0 {
		var s blockState
		if err := dec.Decode(&s); err != nil {
			return nil, fmt.Errorf("state %d: %w", len(out), err)
		}
		out = append(out, s)
	}
	return out, nil
}
