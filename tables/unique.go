package tables

import (
	"sync"
)

type uniqueTableEntry struct {
	mux  sync.Mutex
	vals map[interface{}]*tableEntry
}
