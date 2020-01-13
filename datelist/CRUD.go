package datelist

import (
	"time"
	"github.com/hewiefreeman/GopherDB/helpers"
	"github.com/hewiefreeman/GopherDB/schema"
)



func (d *Datelist) Get(
		start *time.Time, sel []string, where map[string]interface{},
		asc bool, limit int, page int) ([]*DateListEntry, int) {

	// key is required
	if len(key) == 0 {
		return nil, helpers.ErrorKeyRequired
	}

	// default start and end
	if start == nil {
		start = time.Now()
	}

	var e []*DateListEntry


	return e, 0
}