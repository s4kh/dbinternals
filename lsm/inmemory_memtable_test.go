package lsm

import (
	"maps"
	"slices"
	"testing"
)

func TestMemtableInterface(t *testing.T) {
	var _ Memtable = (*InMemoryMemtable)(nil)
}

func TestPutGet(t *testing.T) {
	imm := NewInMemoryMemtable()
	imm.Put("a", "a1")
	want := "a1"
	gs, val := imm.Get("a")
	if gs != FOUND {
		t.Errorf("wanted %s got %s", FOUND, gs)
	}
	if val != want {
		t.Errorf("want %s got %s", want, val)
	}
}

func TestOverwriteGet(t *testing.T) {
	imm := NewInMemoryMemtable()
	imm.Put("a", "a1")
	imm.Put("a", "a2")
	want := "a2"
	gs, val := imm.Get("a")
	if gs != FOUND {
		t.Errorf("wanted %s got %s", FOUND, gs)
	}
	if val != want {
		t.Errorf("want %s got %s", want, val)
	}
}

func TestDeleteGet(t *testing.T) {
	imm := NewInMemoryMemtable()
	imm.Put("a", "a1")
	imm.Delete("a")
	gs, _ := imm.Get("a")
	want := DELETED
	if gs != want {
		t.Errorf("wanted %s got %s", want, gs)
	}

	gs, _ = imm.Get("b")
	if gs != ABSENT {
		t.Errorf("wanted %s got %s", ABSENT, gs)
	}

	t.Run("delete should add tombstone record", func(t *testing.T) {
		oldSize := imm.size
		imm.Delete("b")
		dl, _ := imm.Get("b")
		if dl != DELETED {
			t.Errorf("want %s got %s", DELETED, dl)
		}
		if oldSize == imm.size {
			t.Errorf("old size(%d) == new size(%d)", oldSize, imm.size)
		}
	})
}

func TestDeletePutGet(t *testing.T) {
	imm := NewInMemoryMemtable()
	imm.Put("a", "a1")
	imm.Delete("a")
	gs, _ := imm.Get("a")
	want := DELETED
	if gs != want {
		t.Errorf("wanted %s got %s", want, gs)
	}

	imm.Put("a", "a2")
	want2 := "a2"
	gs, val := imm.Get("a")
	if gs != FOUND {
		t.Errorf("wanted %s got %s", FOUND, gs)
	}
	if val != want2 {
		t.Errorf("wanted %s got %s", want2, val)
	}
}
func TestSeqNum(t *testing.T) {
	imm := NewInMemoryMemtable()
	imm.Put("a", "a")
	imm.Put("b", "b")
	imm.Delete("a")
	imm.Delete("c")
	//seq num in data
	seqNums := []int{3, 2, 4}
	actual := make([]int, 0)
	for _, k := range slices.Sorted(maps.Keys(imm.data)) {
		actual = append(actual, imm.data[k].seqNum)
	}
	if !slices.Equal(seqNums, actual) {
		t.Errorf("seq num wrong; expected:%v, actual:%v", seqNums, actual)
	}
}
func TestIteratorOrder(t *testing.T) {
	imm := NewInMemoryMemtable()
	imm.Put("s", "s1")
	imm.Put("bs", "bs1")
	imm.Put("ca", "ca1")
	imm.Put("a", "a1")
	res := make([]string, 0)
	for k := range imm.Values() {
		res = append(res, k)
	}
	want := []string{"a", "bs", "ca", "s"}
	if !slices.Equal(want, res) {
		t.Errorf("want %v got %v", want, res)
	}
}

func TestSize(t *testing.T) {
	imm := NewInMemoryMemtable()
	shortVal := "s1"
	imm.Put("s", shortVal)
	old := imm.size
	longVal := "s2222222"
	imm.Put("s", longVal)
	oldSize2 := imm.size
	expectedSize := old + int64(len(longVal)) - int64(len(shortVal))
	if expectedSize != oldSize2 {
		t.Errorf("new size %d, not same as expected new size %d", oldSize2, expectedSize)
	}

	imm.Delete("s")
	valSize := len(longVal)
	reducedSize := oldSize2 - int64(valSize)
	if reducedSize != imm.size {
		t.Errorf("delete should reduce old(%d) size, but newSize(%d)", reducedSize, imm.size)
	}
}
