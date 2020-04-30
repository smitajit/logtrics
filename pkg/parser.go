package pkg

import (
	"fmt"
	"regexp"

	"github.com/pkg/errors"
	lua "github.com/yuin/gopher-lua"
)

type (
	// Parser represents the implementation of string parsing
	Parser interface {
		FindSubStrings(s string) (map[string]string, bool)
	}

	// RE2 represents RE2 expression parser
	RE2 struct {
		regexp *regexp.Regexp
	}
)

// NewParser returns a new parser instance
func NewParser(table *lua.LTable) (Parser, error) {
	t := table.RawGet(lua.LString("type")).String()
	switch t {
	case "re2":
		regexp, err := regexp.Compile(table.RawGet(lua.LString("expression")).String())
		if err != nil {
			return nil, errors.Wrap(err, "invalid regular expression")
		}
		return &RE2{regexp: regexp}, nil

	default:
		return nil, fmt.Errorf("parser type not found")
	}

}

// FindSubStrings extracts the sub strings from the string
func (p *RE2) FindSubStrings(s string) (map[string]string, bool) {
	ok := p.regexp.MatchString(s)
	if !ok {
		return nil, false
	}
	subs := make(map[string]string)
	matches := p.regexp.FindStringSubmatch(s)
	n := p.regexp.SubexpNames()
	for i, exp := range n {
		if i >= len(matches) {
			continue
		}
		if exp == "" {
			continue
		}
		subs[exp] = matches[i]
	}
	return subs, true
}
