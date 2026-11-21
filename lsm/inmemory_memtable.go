package lsm

import (
	"iter"
	"maps"
	"slices"
	"unsafe"
)

type GetResult string

const (
	DELETED GetResult = "deleted"
	FOUND   GetResult = "found"
	ABSENT  GetResult = "absent"
)

type Memtable interface {
	Put(key, value string)
	Get(key string) (GetResult, string)
	Delete(key string)
	Values() iter.Seq2[string, *Entry]
}

type Entry struct {
	seqNum      int
	isTombstone bool
	val         string
	key         string
}

func (e *Entry) Size() int64 {
	// unsafe sizeof includes the int and bool
	total := int64(unsafe.Sizeof(*e))
	total += int64(len(e.key))
	total += int64(len(e.val))

	return total
}

type InMemoryMemtable struct {
	data    map[string]*Entry
	counter int
	size    int64
}

func NewInMemoryMemtable() *InMemoryMemtable {
	return &InMemoryMemtable{
		data: make(map[string]*Entry),
	}
}

func (imm *InMemoryMemtable) Values() iter.Seq2[string, *Entry] {
	keys := slices.Collect(maps.Keys(imm.data))
	slices.Sort(keys)

	return func(yield func(string, *Entry) bool) {
		for _, k := range keys {
			val := imm.data[k]
			if !yield(val.key, val) {
				return
			}
		}
	}
}

// Put only keeps the latest version on same key
func (imm *InMemoryMemtable) Put(key, val string) {
	entry, ok := imm.data[key]
	if ok {
		imm.size -= entry.Size()
		imm.counter++
		entry.val = val
		entry.isTombstone = false
		entry.seqNum = imm.counter
		// val is changed so sizes can vary
		imm.size += entry.Size()
		return
	}
	imm.counter++
	entry = &Entry{
		seqNum: imm.counter,
		key:    key,
		val:    val,
	}
	imm.data[key] = entry
	imm.computeSizeForNewEntry(entry)
}

// size = struct + key bytes + val bytes, plus the map's key copy once per distinct key
func (imm *InMemoryMemtable) computeSizeForNewEntry(e *Entry) {
	imm.size += e.Size() + int64(len(e.key))
}

// Get returns FOUND, value; DELETED(tombstone), ""; ABSENT, ""
func (imm *InMemoryMemtable) Get(key string) (GetResult, string) {
	if entry, ok := imm.data[key]; ok {
		if entry.isTombstone {
			return DELETED, ""
		}
		return FOUND, entry.val
	}

	return ABSENT, ""
}

// Delete sets isTombstone flag
// we have to record it even key does not exist
// due to key being in the SSTable
func (imm *InMemoryMemtable) Delete(key string) {
	imm.counter++
	if entry, ok := imm.data[key]; ok {
		imm.size -= entry.Size()
		entry.seqNum = imm.counter
		// saving space
		entry.val = ""
		entry.isTombstone = true
		imm.size += entry.Size()
		return
	}
	// non-existent key might exist in sstable so we need to record it
	tsEntry := &Entry{
		seqNum:      imm.counter,
		isTombstone: true,
		key:         key,
	}
	imm.data[key] = tsEntry
	imm.computeSizeForNewEntry(tsEntry)
}
