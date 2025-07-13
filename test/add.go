package test

import "strings"

func add(a, b int) int {
	return a + b
}

func useStringsBuilder(sb strings.Builder) string {
	sb.WriteString("touch")
	return sb.String()
}

type Adder struct{}

func (adder *Adder) Do(a, b int) int {
	return add(a, b)
}

func callAdder(addr *Adder, a, b int) int {
	return addr.Do(a, b)
}
