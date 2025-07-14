## WIP
### > beware of spaghetti

##  Installation

```bash
go get --tool github.com/moritz-tiesler/boil
```
## Usage 

```bash
$ cat src/frog.go
package frog

import "fmt"

type Frog struct{}

func (f *Frog) quack() string {
	fmt.Println("quack!")
	return "quack!"
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
package frog

import (
	"reflect"
	"testing"
)

func TestQuack(t *testing.T) {
	t.Run("TestQuack_0", func(t *testing.T) {
		// delete this after your implementation
		t.Fatalf("test not implemented")
		receiver := Frog{}

		result0 := receiver.quack()

		var expect0 string
		if !reflect.DeepEqual(result0, expect0) {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}

	})
}
```
