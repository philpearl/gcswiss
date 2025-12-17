package gcswiss

import (
	"hash/maphash"
)

type hashValue uint32

const tableSize = 4096

// Table is a fixed-size hash table containing groups.
type table[K comparable, V any] struct {
	groups [tableSize]group[K, V]

	// localDepth is the number of bits of the hash used to pick this table in
	// the extensible hashing scheme.
	localDepth int
	// used is the number of entries in the table
	used int
	// This is the index of this table in the map's table index.
	index int
}

func newTable[K comparable, V any]() *table[K, V] {
	var t table[K, V]

	for i := range t.groups {
		t.groups[i].init()
	}

	return &t
}

func (t *table[K, V]) clear() {
	for i := range t.groups {
		t.groups[i].init()
	}
	t.used = 0
	t.index = 0
}

var seed = maphash.MakeSeed()

func hash[K comparable](key K) hashValue {
	return hashValue(maphash.Comparable(seed, key))
}

// find looks for the given key in the table, returning its location and whether
// it was found. Values are accessed via the returned GroupLocation.
func (t *table[K, V]) find(key K, hash hashValue) (GroupLocation[K, V], bool) {
	l := hashValue(len(t.groups))

	groupIndex := (hash >> 7) % l
	for range t.groups {
		group := &t.groups[groupIndex]
		matches := group.control.findMatches(hash)
		for matches != 0 {
			index := matches.firstSet()
			entry := &group.entries[index]
			if entry.key == key {
				// Found the key
				return GroupLocation[K, V]{
					table: t,
					group: group,
					index: index,
					hash:  hash,
				}, true
			}
			matches = matches.clearFirstBit()
		}

		if empty := group.control.findEmpty(); empty != 0 {
			// There is an empty slot, so we've reached the end of the probe
			// sequence and the key is not present in the map.
			return GroupLocation[K, V]{
				table: t,
				group: group,
				index: empty.firstSet(),
				hash:  hash,
			}, false
		}
		// Continue to next group in case of hash collision
		// TODO: try a different probe sequence
		groupIndex = (groupIndex + 1) % l
	}
	panic("table is full")
}

func (t *table[K, V]) onSet(m *Map[K, V]) {
	t.used++
	if t.used > len(t.groups)*groupSize*3/4 {
		// Table is too full, need to grow
		m.onGrowthNeeded(t)
	}
}

// split splits the table, returning a new table containing hopefully half of
// the entries.
func (t *table[K, V]) split(m *Map[K, V]) (oldTab, newTab *table[K, V]) {
	// We create two whole new tables rather than reusing the existing one,
	// because we can't enumerate over the old table and modify it at the same
	// time.
	//
	// We do re-use the old table next time we grow - it gets cleared and put
	// into the spare table pool (which has 1 slot!).
	newTab = m.newTable()
	oldTab = m.newTable()
	oldTab.localDepth = t.localDepth + 1
	oldTab.index = t.index * 2
	newTab.localDepth = t.localDepth + 1
	newTab.index = t.index*2 + 1

	// We create a new table, then split the data in the current table between
	// the current table and the new, based on the hash bit that the new local
	// depth exposes.
	mask := hashValue(1 << (32 - t.localDepth - 1))

	for i := range t.groups {
		group := &t.groups[i]
		// Find all the used entries in this group and iterate over them
		matches := group.control.findFull()
		for matches != 0 {
			index := matches.firstSet()
			ent := &group.entries[index]

			// We need to recalculate the hash so that we can find the correct
			// bit to decide what to split. We also need to re-insert the entry
			// into the tables, and the location won't be the same because of
			// probing.
			hash := hash(ent.key)
			tab := oldTab
			if hash&mask != 0 {
				tab = newTab
			}
			loc, ok := tab.find(ent.key, hash)
			if ok {
				panic("found existing key when splitting table")
			}
			loc.Set(ent.key, ent.value)

			matches = matches.clearFirstBit()
		}
	}

	return oldTab, newTab
}
