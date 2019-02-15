package tables

import (
	"sync"
	"github.com/hewiefreeman/GopherGameDB/helpers"
)

type groupedTableEntry struct {
	mux     sync.Mutex
	groups  map[interface{}]*tableEntryGroup
}

type tableEntryGroup struct {
	mux     sync.Mutex
	entries []*tableEntry
}

func (g *groupedTableEntry) getEntryGroup(val interface{}) *tableEntryGroup {
	if !helpers.IsHashable(val) {
		return nil
	}
	g.mux.Lock()
	u := g.groups[val]
	g.mux.Unlock()
	return u
}
