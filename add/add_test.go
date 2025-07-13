package add

import "testing"

func TestAdd(t *testing.T) {
	t.Run("Adding positive numbers", func(t *testing.T) {
		// delete this after your implementation
		t.Fatalf("test not implemented")
		result0 := add(0, 0)

		expect0 := 0
		if result0 != expect0 {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}

	})
}
