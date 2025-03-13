package lib

import "testing"

func TestPanicIfNotValid(t *testing.T) {
	val := CreateNull[string]()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code didn't panic!")
		}
	}()

	val.GetValue()
}
