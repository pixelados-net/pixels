// Command badgeimport generates deterministic PostgreSQL badge reference seed data.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

// entry stores one audited legacy guild editor row.
type entry struct {
	id      int
	valueA  string
	valueB  string
	kind    string
	enabled bool
}

// insertPattern extracts one guilds_elements value tuple.
var insertPattern = regexp.MustCompile(`VALUES \(([0-9]+), '([^']*)', '([^']*)', '([^']*)', '([01])'\);`)

// main loads an audited legacy source and writes one deterministic Liquibase seed.
func main() {
	source := flag.String("source", "", "legacy BaseDB SQL source")
	output := flag.String("output", "", "generated PostgreSQL seed")
	flag.Parse()
	if *source == "" || *output == "" {
		panic("source and output are required")
	}
	entries, err := readEntries(*source)
	if err != nil {
		panic(err)
	}
	if err = writeSeed(*output, entries); err != nil {
		panic(err)
	}
}

// readEntries parses only enabled guild editor reference rows.
func readEntries(path string) ([]entry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	entries := make([]entry, 0, 512)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "INSERT INTO `guilds_elements`") {
			continue
		}
		match := insertPattern.FindStringSubmatch(line)
		if len(match) != 6 {
			return nil, fmt.Errorf("unsupported guild element row: %s", line)
		}
		var id int
		if _, err = fmt.Sscanf(match[1], "%d", &id); err != nil {
			return nil, err
		}
		entries = append(entries, entry{id: id, valueA: match[2], valueB: match[3], kind: match[4], enabled: match[5] == "1"})
	}
	return entries, scanner.Err()
}

// writeSeed writes normalized elements and all three Nitro color collections.
func writeSeed(path string, entries []entry) error {
	order := map[string]int{"base": 0, "symbol": 1, "base_color": 2, "symbol_color": 3, "background_color": 4}
	sort.Slice(entries, func(left, right int) bool {
		if order[entries[left].kind] == order[entries[right].kind] {
			return entries[left].id < entries[right].id
		}
		return order[entries[left].kind] < order[entries[right].kind]
	})
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	defer writer.Flush()
	_, _ = fmt.Fprintln(writer, "--liquibase formatted sql\n\n--changeset pixels:pixels-group-seed-0002-badge-registry")
	for _, item := range entries {
		if !item.enabled {
			continue
		}
		switch item.kind {
		case "base", "symbol":
			kind := 0
			if item.kind == "symbol" {
				kind = 1
			}
			_, _ = fmt.Fprintf(writer, "insert into social_group_badge_elements(kind,id,value_a,value_b,enabled,order_num) values(%d,%d,'%s','%s',true,%d) on conflict(kind,id) do update set value_a=excluded.value_a,value_b=excluded.value_b,enabled=true,order_num=excluded.order_num;\n", kind, item.id, sqlValue(item.valueA), sqlValue(item.valueB), item.id)
		case "base_color", "symbol_color", "background_color":
			family := order[item.kind] - 2
			_, _ = fmt.Fprintf(writer, "insert into social_group_badge_colors(family,id,hex,enabled,order_num) values(%d,%d,'%s',true,%d) on conflict(family,id) do update set hex=excluded.hex,enabled=true,order_num=excluded.order_num;\n", family, item.id, strings.ToUpper(item.valueA), item.id)
		default:
			return fmt.Errorf("unsupported guild element kind %q", item.kind)
		}
	}
	_, _ = fmt.Fprintln(writer, "\n--rollback delete from social_group_badge_colors; delete from social_group_badge_elements;")
	return nil
}

// sqlValue escapes one trusted manifest string for a single-quoted literal.
func sqlValue(value string) string { return strings.ReplaceAll(value, "'", "''") }
