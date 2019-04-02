package main

import (

)

type databaseEntry struct {
	key  string
	mmr  int
	rank int
}

var (
	db1 []databaseEntry = []databaseEntry{
			databaseEntry{key: "Joe", mmr: 1456, rank: 4},
			databaseEntry{key: "Marg", mmr: 1435, rank: 3},
			databaseEntry{key: "Barney", mmr: 1622, rank: 6},
			databaseEntry{key: "Moe", mmr: 1634, rank: 6},
			databaseEntry{key: "Bob", mmr: 1512, rank: 5},
			databaseEntry{key: "Sally", mmr: 1765, rank: 8},
			databaseEntry{key: "Ward", mmr: 1692, rank: 7},
			databaseEntry{key: "Ugly", mmr: 1120, rank: 1},
			databaseEntry{key: "Dumb", mmr: 1253, rank: 3},
			databaseEntry{key: "Quark", mmr: 1923, rank: 9}}

	db2 []databaseEntry = []databaseEntry{
			databaseEntry{key: "Doge", mmr: 1654, rank: 6},
			databaseEntry{key: "Weeb", mmr: 1556, rank: 5},
			databaseEntry{key: "Knack", mmr: 1745, rank: 8},
			databaseEntry{key: "You", mmr: 1214, rank: 2},
			databaseEntry{key: "BAMF", mmr: 1398, rank: 4},
			databaseEntry{key: "Will", mmr: 1322, rank: 3},
			databaseEntry{key: "Tony", mmr: 1310, rank: 3},
			databaseEntry{key: "Bam", mmr: 1564, rank: 5},
			databaseEntry{key: "Mary", mmr: 1089, rank: 1},
			databaseEntry{key: "Dane", mmr: 1045, rank: 1}}

	db3 []databaseEntry = []databaseEntry{
			databaseEntry{key: "Rager", mmr: 1785, rank: 8},
			databaseEntry{key: "Beautiful", mmr: 1712, rank: 7},
			databaseEntry{key: "Sony", mmr: 1346, rank: 4},
			databaseEntry{key: "Microsoft", mmr: 1474, rank: 5},
			databaseEntry{key: "Ubisoft", mmr: 1000, rank: 1},
			databaseEntry{key: "Nintendo", mmr: 2250, rank: 10},
			databaseEntry{key: "Sega", mmr: 1521, rank: 5},
			databaseEntry{key: "Capcom", mmr: 1511, rank: 5},
			databaseEntry{key: "Marvel", mmr: 1478, rank: 4},
			databaseEntry{key: "Torvalds", mmr: 1590, rank: 6}}
)

type databaseSearch struct {
	db []databaseEntry
	totalLen int

	// Numeric searches
	min int
	max int

	// Alphanumeric searches
	min string
	max string
}

func main() {
	// Paginate through databases with page/limit, sorted by mmr
	getPage(1, 5, [][]databaseEntry{dbl, db2, db3});
}

func getPage(page int, limit int, dbl [][]databaseEntry) {
	// Get total length of search and sorted items for each database
	var dbs []databaseSearch = make([]databaseSearch, len(dbl), len(dbl))
	for i := 0; i < len(dbl); i++ {
		// This should be stored on each shard and use a key to continue query (with a 3 second timeout)
		dbs[i] = getPageInit(dbl[i], page, limit, 1000, 2000)
	}

	//

}

// Get
func getPageInit(db []databaseEntry, page int, limit int, mmrMin int, mmrMax int) databaseSearch {
	dbs := databaseSearch{totalLen: len(db)}
	// Get entries with mmr between mmrMin and mmrMax, simultaneously sorting
	for i := 0; i < len(db); i++ {
		if db[i].mmr >= mmrMin && db[i].mmr <= mmrMax {
			// Insert into array at sorted position

		}

		// Reached max entries for search
		if len(dbs.db) == ((page-1)*limit)+limit {
			return dbs
		}
	}
}
