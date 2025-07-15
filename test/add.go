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

type Adder struct{}

func (adder *Adder) Do(a, b int) int {
	return add(a, b)
}

func createAdder() *Adder {
	return &Adder{}
}

func callAdder(addr *Adder, a, b int) int {
	return addr.Do(a, b)
}
