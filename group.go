package gcswiss

const groupSize = 8

type group[K comparable, V any] struct {
	control groupControl
	entries [groupSize]entry[K, V]
}

type entry[K comparable, V any] struct {
	key   K
	value V
}

type groupControl uint64

const (
	emptyControl       byte         = 0x80
	emptyGroupControl  groupControl = 0x8080808080808080
	controlHashMask                 = 0x7F
	groupControlExpand              = 0x0101010101010101
)

func (g *group[K, V]) init() {
	g.control = emptyGroupControl
}

// findMatches returns a bits mask of which entries in the group match the given
// hash value.
func (gc groupControl) findMatches(hash hashValue) groupBits {
	ctrlHash := byte(hash & controlHashMask)
	// Find the entries where the control byte matches ctrlHash
	//
	// We expand the ctrlHash to a groupControl where each byte is ctrlHash,
	// then XOR that with the group control. Any byte that was equal will now be
	// zero. We then subtract 0x01 from each byte, so any byte that was zero
	// will now have its high bit set. Finally we AND with 0x80 to keep only the
	// high bits.
	//
	// Note this does give false positives!
	matchesAreZero := uint64(gc) ^ (uint64(ctrlHash) * groupControlExpand)
	return groupBits(((matchesAreZero - 0x0101010101010101) &^ matchesAreZero) & 0x8080808080808080)
}

// findEmpty returns a bits mask of which entries in the group are empty.
func (gc groupControl) findEmpty() groupBits {
	return groupBits(uint64(gc) & uint64(emptyGroupControl))
}

func (gc groupControl) findFull() groupBits {
	return groupBits(^uint64(gc) & uint64(emptyGroupControl))
}

// GroupLocation represents a location in the table for a specific key. The
// caller retrieves the location, then can use it to get or set the value.
// If the key did not exist, the caller must use Set to set the key and value.
type GroupLocation[K comparable, V any] struct {
	m     *Map[K, V]
	table *table[K, V]
	group *group[K, V]
	index int
	hash  hashValue
}

func (gl GroupLocation[K, V]) Set(key K, value V) {
	gl.group.entries[gl.index] = entry[K, V]{key: key, value: value}
	gl.group.control = (gl.group.control &^ (groupControl(0x80) << (gl.index * 8))) | groupControl(byte(gl.hash&0x7F))<<(gl.index*8)
	gl.table.onSet(gl.m)
}

func (gl GroupLocation[K, V]) SetValue(value V) {
	gl.group.entries[gl.index].value = value
}

func (gl GroupLocation[K, V]) Get() V {
	return gl.group.entries[gl.index].value
}
