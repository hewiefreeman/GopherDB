package tables

import (
	"sync"
	"github.com/hewiefreeman/GopherGameDB/helpers"
)

type UniqueTableItem struct {
	mux  sync.Mutex
	vals map[interface{}]*TableEntry
}

func (u *UniqueTableItem) getEntry(val interface{}) *TableEntry {
	if !helpers.IsHashable(val) {
		return nil
	}
	u.mux.Lock()
	e := u.vals[val]
	u.mux.Unlock()
	return e
}

func (u *UniqueTableItem) addEntry(val interface{}) {

}
