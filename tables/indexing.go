package ggdb

import (
	"time"
	"github.com/hewiefreeman/GopherGameDB/helpers"
)

/////////////////////////////////////////////////////////////////////////////////////////////////
//   Indexing   /////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////

func (t *table) optimizeIndex() {
	(*t).iMux.Lock()
	(*t).eMux.Lock()
	(*t).index = make([]*tableEntry, len((*t).entries), len((*t).entries))
	i := 0
	for _, e := range (*t).entries {
		(*t).index[i] = e
		i++
	}
	(*t).optIMux.Lock()
	(*t).lastOptIndex = time.Now()
	(*t).optIMux.Unlock()
	(*t).eMux.Unlock()
	(*t).iMux.Unlock()
}

func (t *table) lastIndexOptimizationTime() time.Time {
	(*t).optIMux.Lock()
	i := (*t).lastOptIndex
	(*t).optIMux.Unlock()
	return i
}

func (t *table) addToIndex(v *tableEntry) {
	(*t).iMux.Lock()
	(*t).index = append((*t).index, v)
	(*t).iMux.Unlock()
}

func (t *table) removeFromIndex(v *tableEntry) {
	(*t).iMux.Lock()
	for i := 0; i < len((*t).index); i++ {
		if (*t).index[i] == v {
			(*t).index[i] = nil
		}
	}
	(*t).iMux.Unlock()
}

func (t *table) indexSize() int {
	(*t).iMux.Lock()
	s := len((*t).index)
	(*t).iMux.Unlock()
	return s
}

func (t *table) indexChunk(start int, amount int) ([]*tableEntry, int) {
	if start < 0 || amount <= 0 {
		return []*tableEntry{}, helpers.ErrorIndexChunkOutOfRange
	}
	max := start+amount;

	(*t).iMux.Lock()
	indexLen := len((*t).index)
	if start >= indexLen {
		start = indexLen-1
		max = indexLen
	} else if max > indexLen {
		max = indexLen
	}
	c := (*t).index[start:max]
	(*t).iMux.Unlock()

	return c, 0
}
