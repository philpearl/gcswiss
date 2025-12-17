package gcswiss

import (
	"strconv"
	"testing"
)

func TestTableSet(t *testing.T) {
	tab := newTable[int]()

	for i := range 10000 {
		key := strconv.Itoa(i)
		loc, found := tab.find(key, hash(key))
		if found {
			t.Errorf("expected key %d to not be found", i)
		}
		loc.Set(strconv.Itoa(i), i)
	}

	for i := range 10000 {
		key := strconv.Itoa(i)
		loc, found := tab.find(key, hash(key))
		if !found {
			t.Errorf("expected key %d to be found", i)
		}
		actual := loc.Get()
		if actual != i {
			t.Errorf("expected value %d, got %d", i, actual)
		}
	}
}
