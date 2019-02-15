package tables

import (
	"sync"
	"github.com/hewiefreeman/GopherGameDB/helpers"
)

type uniqueTableEntry struct {
	mux  sync.Mutex
	vals map[interface{}]*tableEntry
}

func (u *uniqueTableEntry) getEntry(val interface{}) *tableEntry {
	if !helpers.IsHashable(val) {
		return nil
	}
	u.mux.Lock()
	e := u.vals[val]
	u.mux.Unlock()
	return e
}
