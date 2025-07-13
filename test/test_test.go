package test

import "testing"
import "strings"
import "github.com/moritz-tiesler/fmock"


func TestAdd(t *testing.T) {
	t.Run("TestAdd_0", func(t *testing.T) {
		// delete this after your implementation
		t.Fatalf("test not implemented")
		
		result0 := add(0, 0)
		
		expect0 := 0
		if result0 != expect0 {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}
	
	})
}

func TestUseStringsBuilder(t *testing.T) {
	t.Run("TestUseStringsBuilder_0", func(t *testing.T) {
		// delete this after your implementation
		t.Fatalf("test not implemented")
		
		result0 := useStringsBuilder(strings.Builder{})
		
		expect0 := ""
		if result0 != expect0 {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}
	
	})
}

func TestUseFmock(t *testing.T) {
	t.Run("TestUseFmock_0", func(t *testing.T) {
		// delete this after your implementation
		t.Fatalf("test not implemented")
		
		result0 := useFmock(fmock.Mock{})
		
		expect0 := false
		if result0 != expect0 {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}
	
	})
}

func TestDo(t *testing.T) {
	t.Run("TestDo_0", func(t *testing.T) {
		// delete this after your implementation
		t.Fatalf("test not implemented")
		receiver := Adder{}
		result0 := receiver.Do(0, 0)
		
		expect0 := 0
		if result0 != expect0 {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}
	
	})
}

func TestCallAdder(t *testing.T) {
	t.Run("TestCallAdder_0", func(t *testing.T) {
		// delete this after your implementation
		t.Fatalf("test not implemented")
		
		result0 := callAdder(nil, 0, 0)
		
		expect0 := 0
		if result0 != expect0 {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}
	
	})
}
