package test

import "testing"


func TestADD(t *testing.T) {
	t.Run("TestADD_0", func(t *testing.T) {
		// delete this after your implementation
		t.Fatalf("test not implemented")
		
		result0 := add(0, 0)
		
		expect0 := 0
		if result0 != expect0 {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}
	
	})
}

func TestDO(t *testing.T) {
	t.Run("TestDO_0", func(t *testing.T) {
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

func TestCALLADDER(t *testing.T) {
	t.Run("TestCALLADDER_0", func(t *testing.T) {
		// delete this after your implementation
		t.Fatalf("test not implemented")
		
		result0 := callAdder(nil, 0, 0)
		
		expect0 := 0
		if result0 != expect0 {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}
	
	})
}
