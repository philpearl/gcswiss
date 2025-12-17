package gcswiss

import (
	"strconv"
	"testing"
)

func TestSetGet(t *testing.T) {
	m := New[string, int]()

	keys := make([]string, 1_000_000)
	for i := range keys {
		keys[i] = "key" + strconv.Itoa(i)
	}

	for i, key := range keys {
		loc, ok := m.Find(key)
		if ok {
			t.Fatalf("expected key %s to not be found", key)
		}
		loc.Set(key, i)
	}

	for i, key := range keys {
		loc, found := m.Find(key)
		if !found {
			t.Fatalf("expected key %s to be found", key)
		}
		value := loc.Get()
		if value != i {
			t.Fatalf("expected value %d for key %s, got %d", i, key, value)
		}
	}
}

func TestSetGetUpdate(t *testing.T) {
	m := New[string, int]()

	keys := make([]string, 1_000_000)
	for i := range keys {
		keys[i] = "key" + strconv.Itoa(i)
	}

	for i, key := range keys {
		loc, ok := m.Find(key)
		if ok {
			t.Fatalf("expected key %s to not be found", key)
		}
		loc.Set(key, i)
	}

	for i, key := range keys {
		loc, found := m.Find(key)
		if !found {
			t.Fatalf("expected key %s to be found", key)
		}
		value := loc.Get()
		if value != i {
			t.Fatalf("expected value %d for key %s, got %d", i, key, value)
		}
		loc.SetValue(i * 10)
	}

	for i, key := range keys {
		loc, found := m.Find(key)
		if !found {
			t.Fatalf("expected key %s to be found", key)
		}
		value := loc.Get()
		if value != i*10 {
			t.Fatalf("expected value %d for key %s, got %d", i*10, key, value)
		}
	}
}

func BenchmarkSetGet(b *testing.B) {
	keys := make([]string, 10_000_000)
	for i := range keys {
		keys[i] = "key" + strconv.Itoa(i)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		m := New[string, int]()
		for i, key := range keys {
			loc, ok := m.Find(key)
			if ok {
				b.Fatalf("expected key %s to not be found", key)
			}
			loc.Set(key, i)
		}

		for i, key := range keys {
			loc, found := m.Find(key)
			if !found {
				b.Fatalf("expected key %s to be found", key)
			}
			value := loc.Get()
			if value != i {
				b.Fatalf("expected value %d for key %s, got %d", i, key, value)
			}
		}
	}
}

func BenchmarkSetGetMap(b *testing.B) {
	keys := make([]string, 10_000_000)
	for i := range keys {
		keys[i] = "key" + strconv.Itoa(i)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		m := make(map[string]int, 4096*8)
		for i, key := range keys {
			m[key] = i
		}

		for i, key := range keys {
			value, ok := m[key]
			if !ok {
				b.Fatalf("expected key %s to be found", key)
			}
			if value != i {
				b.Fatalf("expected value %d for key %s, got %d", i, key, value)
			}
		}
	}
}
