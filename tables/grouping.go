package tables

import (
	"sync"
	"github.com/hewiefreeman/GopherGameDB/helpers"
)

type GroupedTableEntry struct {
	mux     sync.Mutex
	groups  map[interface{}]*TableEntryGroup
}

type TableEntryGroup struct {
	mux     sync.Mutex
	entries []*TableEntry
}

func (g *GroupedTableEntry) Get(val interface{}) *TableEntryGroup {
	if !helpers.IsHashable(val) {
		return nil
	}
	g.mux.Lock()
	u := g.groups[val]
	g.mux.Unlock()
	return u
}

func (t *TableEntryGroup) Len() int {
	t.mux.Lock()
	l := len(t.entries)
	t.mux.Unlock()
	return l
}

func (t *TableEntryGroup) Get(i int) *TableEntry {
	t.mux.Lock()
	if i < 0 || i >= len(t.entries) {
		t.mux.Unlock()
		return nil
	}
	e := t.entries[i]
	t.mux.Unlock()
	return e
}
