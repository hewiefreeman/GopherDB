package tables

import (

)

type tableSchema struct {
	items map[string]tableSchemaItem
}

type tableSchemaItem struct {
	index       int
	unique      bool
	grouped     bool
	groupedVals []interface{}
}

func newTableSchema(items []string, unique []int, grouped map[int][]interface{}) (tableSchema, int) {
	if len(items) == 0 {
		return tableSchema{}, helpers.ErrorSchemaItemsRequired
	}
	var s tableSchema
	s.items = make(map[string]tableSchemaItem)
	// Go through items
	for i := 0; i < len(items); i++ {
		se := tableSchemaItem{index: i}
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
						return tableSchema{}, helpers.ErrorUnhashableGroupValue
					} else if v[j] == nil {
						return tableSchema{}, helpers.ErrorNilGroupValue
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

func (s tableSchema) getItem(n string) (tableSchemaItem, int) {
	if v, ok := s.items[n]; ok {
		return v, 0
	}
	return tableSchemaItem{}, helpers.ErrorSchemaItemDoesntExist
}

func (s tableSchemaItem) isUnique(n string) bool {
	return v.unique
}

func (s tableSchemaItem) isGrouped(n string) bool {
	if v, ok := s.items[n]; ok {
		return v.grouped
	}
	return false
}

func (s tableSchemaItem) isGroupedVal(v interface{}) bool {
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
