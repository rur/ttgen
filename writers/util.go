package writers

import (
	"fmt"
	"regexp"
	"sort"

	generate "github.com/rur/ttgen"
)

func SanitizeName(name string) (string, error) {
	re := regexp.MustCompile("(?i)^[a-z][a-z0-9-_]*$")
	if !re.MatchString(name) {
		return name, fmt.Errorf("Invalid name '%s'", name)
	}
	return generate.ValidIdentifier(name), nil
}

type blockDef struct {
	name     string
	ident    string
	partials []*generate.PartialDef
}

func IterateSortedBlocks(blocks map[string][]*generate.PartialDef) ([]blockDef, error) {
	output := make([]blockDef, 0, len(blocks))
	var keys []string
	for k := range blocks {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		ident, err := SanitizeName(k)
		if err != nil {
			return output, fmt.Errorf("Invalid block name '%s'", k)
		}
		output = append(output, blockDef{
			name:     k,
			ident:    ident,
			partials: blocks[k],
		})
	}
	return output, nil
}

// produce a slice of views sorted lex' by their view name
func IterateSortedViews(views map[string]*generate.PartialDef) []*generate.PartialDef {
	output := make([]*generate.PartialDef, 0, len(views))
	var keys []string
	for k := range views {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		output = append(output, views[k])
	}
	return output
}
