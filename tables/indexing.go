package tables

import (
	"sync"
	"github.com/hewiefreeman/GopherGameDB/helpers"
)

type indexChunk struct {
	mux     sync.Mutex
	entries []*TableEntry
}

func (t *Table) getIndexChunk(i int) *indexChunk {
	if i < 0 || i > t.indexChunks-1 {
		return nil
	}
	t.iMux.Lock()
	chunk := t.index[i]
	t.iMux.Unlock()
	return chunk
}

func (t *Table) getIndexChunkForKey(key string) *indexChunk {
	index := helpers.HashString(key)%t.indexChunks
	t.iMux.Lock()
	chunk := t.index[index]
	t.iMux.Unlock()
	return chunk
}

func (t *Table) addToIndex(key string, entry *TableEntry) {
	if entry == nil {
		return
	}

	chunk := t.getIndexChunkForKey(key)

	chunk.mux.Lock()
	chunk.entries = append(chunk.entries, entry)
	chunk.mux.Unlock()
}

func (t *Table) removeFromIndex(key string, entry *TableEntry) {
	if entry == nil {
		return
	}

	chunk := t.getIndexChunkForKey(key)

	chunk.mux.Lock()
	for i := 0; i < len(chunk.entries); i++ {
		if chunk.entries[i] == entry {
			chunk.entries = append(chunk.entries[:i], chunk.entries[i+1:]...)
			break
		}
	}
	chunk.mux.Unlock()
}
