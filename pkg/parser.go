package pkg

import (
	"regexp"

	"github.com/pkg/errors"
)

type (
	// Parser represents the implementation of string parsing
	Parser struct {
		re     string
		regexp *regexp.Regexp
	}
)

// NewParser returns a new parser instance
func NewParser(re string) (*Parser, error) {
	regexp, err := regexp.Compile(re)
	if err != nil {
		return nil, errors.Wrap(err, "invalid regular expression")
	}
	return &Parser{re: re, regexp: regexp}, nil
}

// FindSubStrings extracts the sub strings from the string
func (p *Parser) FindSubStrings(s string) ([]string, bool) {
	ok := p.regexp.MatchString(s)
	if !ok {
		return nil, false
	}
	subs := make([]string, 0)
	matches := p.regexp.FindStringSubmatch(s)
	n := p.regexp.SubexpNames()
	for i, exp := range n {
		if i >= len(matches) {
			continue
		}
		if exp == "" {
			continue
		}
		subs = append(subs, matches[i])
	}
	return subs, true
}
