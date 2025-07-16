# boil
Generate tests for a go package 

  

##  Installation

```bash
go get --tool github.com/moritz-tiesler/boil
```
## Example 

```bash
$ cat src/frog.go
```
```go
package frog

type Frog struct{}

func (f *Frog) quack(loud bool) string {
	if loud {
		return "QUACK!"
	}
	return "quack..."
}
```
```bash
$ cd src/
$ go tool boil
Created file frog_test.go
Created 1 tests
```

```bash
$ cat frog_test.go
```
```go
package frog

import (
	"reflect"
	"testing"
)

func TestQuack(t *testing.T) {
	t.Run("TestQuack_0", func(t *testing.T) {

		// delete this after your implementation
		t.Fatalf("test not implemented")

		var receiver Frog
		var loud bool

		result0 := receiver.quack(loud)

		var expect0 string
		if !reflect.DeepEqual(result0, expect0) {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}

	})
}
```
### Table Tests

You can generate your tests in table format with the `--table` flag:

```bash
$ go tool boil --table
```
