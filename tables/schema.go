package tables

import (
	"github.com/hewiefreeman/GopherGameDB/helpers"
)

type TableSchema struct {
	items map[string]TableSchemaItem
}

type TableSchemaItem struct {
	index       int
	unique      bool
	grouped     bool
	groupedVals []interface{}
}

func NewTableSchema(items []string, unique []int, grouped map[int][]interface{}) (TableSchema, int) {
	if len(items) == 0 {
		return TableSchema{}, helpers.ErrorSchemaItemsRequired
	}
	var s TableSchema
	s.items = make(map[string]TableSchemaItem)
	// Go through items
	for i := 0; i < len(items); i++ {
		se := TableSchemaItem{index: i}
		// check for unique
		for j := 0; j < len(unique); j++ {
			if unique[j] == i {
				se.unique = true
				break
			}
		}
		// check for grouped
		if !se.unique {
			if v, ok := grouped[i]; ok {
				// check if any of the values aren't hashable
				for j := 0; j < len(v); j++ {
					if !helpers.IsHashable(v[j]) {
						return TableSchema{}, helpers.ErrorUnhashableGroupValue
					} else if v[j] == nil {
						return TableSchema{}, helpers.ErrorNilGroupValue
					}
				}
				se.grouped = true
				se.groupedVals = v
			}
		}
		s.items[items[i]] = se
	}
	return s, 0
}

func (s TableSchema) Get(n string) (TableSchemaItem, int) {
	if v, ok := s.items[n]; ok {
		return v, 0
	}
	return TableSchemaItem{}, helpers.ErrorSchemaItemDoesntExist
}

func (s TableSchemaItem) IsUnique(n string) bool {
	return v.unique
}

func (s TableSchemaItem) IsGrouped(n string) bool {
	if v, ok := s.items[n]; ok {
		return v.grouped
	}
	return false
}

func (s TableSchemaItem) IsGroupedVal(v interface{}) bool {
	if !helpers.IsHashable(v) {
		return false
	}
	for i := 0; i < len(s.groupedVals); i++ {
		if s.groupedVals[i] == v {
			return true
		}
	}
	return false
}
