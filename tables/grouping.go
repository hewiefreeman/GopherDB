package tables

import (
	"sync"
)

type groupedTableEntry struct {
	mux     sync.Mutex
	groups  map[interface{}]*tableEntryGroup
}

type tableEntryGroup struct {
	mux     sync.Mutex
	entries []*tableEntry
}
