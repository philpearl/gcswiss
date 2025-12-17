// Package gcswiss is a GC friendly hash map that uses an Extensible hashing
// plus Swiss table design. It is very based on the newer Go standard library
// map and the cockroachdb implementation.
//
// - It does not support deletes. - It uses a 32bit hash - It allocates memory
// outside of the Go heap.
//
// At the bottom we have groups of 8 (maybe 16 later?) entries.
//
// Above this we have tables, which are fixed-size hash tables containing
// groups.
//
// Above that we have a directory, which is indexed by the top bits of the hash
// and points to tables. This is the extensible hashing part.
package gcswiss

import (
	"unsafe"

	"github.com/philpearl/mmap"
	stringbank "github.com/philpearl/stringbank/offheap"
)

type Map[V any] struct {
	tables []*table[V]

	tableIndexShift int
	spareTable      *table[V]
	sb              stringbank.Stringbank
}

func New[V any]() *Map[V] {
	m := Map[V]{
		tableIndexShift: 32,
	}

	var err error
	m.tables, err = mmap.Alloc[*table[V]](1)
	if err != nil {
		panic(err)
	}
	m.tables[0] = m.newTable()

	return &m
}

func (m *Map[V]) Close() {
	m.sb.Close()
	for _, t := range m.tables {
		m.freeTable(t)
	}
	if m.spareTable != nil {
		m.freeTable(m.spareTable)
	}
	mmap.Free(m.tables)
}

func (m *Map[V]) Find(key string) (GroupLocation[V], bool) {
	hash := hash(key)
	tableIndex := (hash >> uint32(m.tableIndexShift))
	table := m.tables[tableIndex]
	return table.find(m, key, hash)
}

func (m *Map[V]) newTable() *table[V] {
	if m.spareTable != nil {
		t := m.spareTable
		m.spareTable = nil
		return t
	}
	tables, err := mmap.Alloc[table[V]](1)
	if err != nil {
		panic(err)
	}
	tables[0].init()
	return &tables[0]
}

func (m *Map[V]) freeTable(t *table[V]) {
	if m.spareTable == nil {
		t.init()
		m.spareTable = t
		return
	}
	if err := mmap.Free(unsafe.Slice(t, 1)); err != nil {
		panic(err)
	}
}

// This is called when a table detects it is too full and needs to grow.
func (m *Map[V]) onGrowthNeeded(t *table[V]) {
	globalDepth := 32 - m.tableIndexShift
	if t.localDepth == globalDepth {
		// Need to grow the directory. This will take care of splitting tables as needed.
		m.grow()
		globalDepth++
	}

	// There should be a relationship between index and depth, and we need to update index when local depth changes
	// 0 0 0 0 0 0
	//   1 0 0 0 1
	//.    2 0 1 2
	//.    3 0 1 3
	//       2 2 4
	//       2 2 5
	//.      3 3 6
	//.      3 3 7

	// We can just split this table, and split up the slots it is currently
	// installed in in the directory.
	oldTab, newTab := t.split(m)
	m.insertTable(oldTab)
	m.insertTable(newTab)
	m.freeTable(t)
}

func (m *Map[V]) insertTable(t *table[V]) {
	depthDifference := 32 - m.tableIndexShift - t.localDepth
	index := t.index * (depthDifference + 1)
	tableWidth := 1 << depthDifference
	for i := range tableWidth {
		m.tables[index+i] = t
	}
}

// grow grows the map by splitting tables as needed. We always double the number
// of entries in the table index, but only split tables as needed. If we don't
// need to split a table we double the number of entries that point to the same
// table.
func (m *Map[V]) grow() {
	newTables, err := mmap.Alloc[*table[V]](len(m.tables) * 2)
	if err != nil {
		panic(err)
	}
	for i, table := range m.tables {
		newTables[i*2] = table
		newTables[i*2+1] = table
	}
	m.tableIndexShift--
	mmap.Free(m.tables)
	m.tables = newTables
}
