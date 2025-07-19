package test

import (
	"strings"

	"github.com/moritz-tiesler/fmock"
)

func add(a, b int) int {
	return a + b
}

func useStringsBuilder(sb strings.Builder) string {
	sb.WriteString("touch")
	return sb.String()
}

func useFmock(a fmock.Mock) bool {
	return a.Called()
}

type Adder struct {
	offset int
}

func (adder *Adder) Do(a, b int) int {
	return add(a, b)
}

func createAdder() *Adder {
	return &Adder{}
}

func callAdder(addr *Adder, a, b int) int {
	return addr.Do(a, b)
}

func NewAdder(offset int) *Adder {
	return &Adder{offset: offset}
}

func NewAdders(offset int) []*Adder {
	var adders []*Adder
	adders = append(adders, &Adder{offset: offset})
	return adders
}

func NewAddersMap(offset int) map[int]*Adder {
	adders := make(map[int]*Adder)
	for i := range offset {
		adders[i] = &Adder{offset: i}
	}
	return adders
}
