package gcswiss

import (
	"fmt"
	"runtime"
	"strconv"
	"testing"
	"unsafe"

	stringbank "github.com/philpearl/stringbank/offheap"
)

func TestSetGet(t *testing.T) {
	m := New[int]()
	defer m.Close()

	var sb stringbank.Stringbank
	var buf []byte
	for i := range 1_000_000 {
		buf = fmt.Appendf(buf[:0], "key%d", i)
		sb.Save(unsafe.String(&buf[0], len(buf)))
	}

	i := 0
	for key := range sb.All() {
		loc, ok := m.Find(key)
		if ok {
			t.Fatalf("expected key %s to not be found", key)
		}
		loc.Set(key, i)
		i++
	}

	i = 0
	for key := range sb.All() {
		loc, found := m.Find(key)
		if !found {
			t.Fatalf("expected key %s to be found", key)
		}
		value := loc.Get()
		if value != i {
			t.Fatalf("expected value %d for key %s, got %d", i, key, value)
		}
		i++
	}
}

func TestSetGetUpdate(t *testing.T) {
	m := New[int]()
	defer m.Close()

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
	var sb stringbank.Stringbank
	var buf []byte
	for i := range 10_000_000 {
		buf = append(buf[:0], "key"...)
		buf = strconv.AppendInt(buf, int64(i), 10)
		sb.Save(unsafe.String(&buf[0], len(buf)))
	}

	var i int
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		m := New[int]()
		i = 0
		for key := range sb.All() {
			loc, ok := m.Find(key)
			if ok {
				b.Fatalf("expected key %s to not be found", key)
			}
			loc.Set(key, i)
			i++
		}

		i = 0
		for key := range sb.All() {
			loc, found := m.Find(key)
			if !found {
				b.Fatalf("expected key %s to be found", key)
			}
			value := loc.Get()
			if value != i {
				b.Fatalf("unexpected for key %s, got %d", key, value)
			}
			i++
		}
		m.Close()
		runtime.GC()
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
		runtime.GC()
	}
}
