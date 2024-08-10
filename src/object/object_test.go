package object

import "testing"

func TestStringHashKey(t *testing.T) {
	hello1 := &String{Value: "hello"}
	hello2 := &String{Value: "hello"}

	diff1 := &String{Value: "mahi"}
	diff2 := &String{Value: "mahi"}

	if hello1.HashKey() != hello2.HashKey() {
		t.Errorf("string with same content have different hash keys")
	}

	if diff1.HashKey() != diff2.HashKey() {
		t.Errorf("string with same content have different hash keys")
	}

	if hello1.HashKey() == diff1.HashKey() {
		t.Errorf("string with different content have same hash keys")
	}
}
