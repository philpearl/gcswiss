package gcswiss

import (
	"strconv"
	"testing"
)

func TestTableSet(t *testing.T) {
	var tab table[int]
	tab.init()
	m := New[int]()
	defer m.Close()

	for i := range 10000 {
		key := strconv.Itoa(i)
		loc, found := tab.find(m, key, hash(key))
		if found {
			t.Errorf("expected key %d to not be found", i)
		}
		loc.Set(strconv.Itoa(i), i)
	}

	for i := range 10000 {
		key := strconv.Itoa(i)
		loc, found := tab.find(m, key, hash(key))
		if !found {
			t.Errorf("expected key %d to be found", i)
		}
		actual := loc.Get()
		if actual != i {
			t.Errorf("expected value %d, got %d", i, actual)
		}
	}
}
