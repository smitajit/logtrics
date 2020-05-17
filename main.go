package main

import (
	"fmt"
	"strings"

	"github.com/smitajit/logtrics/pkg"
)

func main() {
	p, err := pkg.NewParser(``)
	if err != nil {
		panic(err)
	}

	ss, matched := p.FindSubStrings(`hello "world"`)
	if !matched {
		panic("not matched")
	}

	fmt.Println(strings.Join(ss, "---"))
}
