package tables

import (
	"sync"
	"github.com/hewiefreeman/GopherGameDB/helpers"
)

type UniqueTableItem struct {
	mux  sync.Mutex
	vals map[interface{}]*TableEntry
}

func (u *UniqueTableItem) getEntry(val interface{}) (*TableEntry, int) {
	if !helpers.IsHashable(val) {
		return nil, helpers.ErrorUnhashableUniqueValue
	}
	u.mux.Lock()
	e := u.vals[val]
	u.mux.Unlock()
	return e, 0
}

func (u *UniqueTableItem) addEntry(val interface{}, entry *TableEntry) int {
	if !helpers.IsHashable(val) {
		return helpers.ErrorUnhashableUniqueValue
	}
	u.mux.Lock()
	if u.vals[val] != nil {
		u.mux.Unlock()
		return helpers.ErrorUniqueValueInUse
	}
	u.vals[val] = entry
	u.mux.Unlock()

	return 0
}

func (u *UniqueTableItem) updateEntry(oldVal interface{}, newVal interface{}, entry *TableEntry) int {
	if !helpers.IsHashable(oldVal) || !helpers.IsHashable(newVal) {
		return helpers.ErrorUnhashableUniqueValue
	}
	u.mux.Lock()
	if u.vals[newVal] != nil {
		u.mux.Unlock()
		return helpers.ErrorUniqueValueInUse
	}
	u.vals[newVal] = entry
	delete(u.vals, oldVal)
	u.mux.Unlock()
}
