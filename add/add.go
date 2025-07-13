package add

func add(a, b int) int {
	return a + b
}

type Adder struct{}

func (adder *Adder) Do(a, b int) int {
	return add(a, b)
}

func callAdder(addr *Adder, a, b int) int {
	return addr.Do(a, b)
}
