package leaderboard

import (
	"fmt"
	"github.com/hewiefreeman/GopherDB/helpers"
	"sync"
)

// Potential memory leak problem? Reference: https://github.com/golang/go/wiki/SliceTricks#delete-without-preserving-order

var (
	leaderboardsMux sync.Mutex
	leaderboards    map[string]*Leaderboard = make(map[string]*Leaderboard)
)

type Leaderboard struct {
	maxEntries    uint64
	dupePushAbove bool
	alwaysReplace bool

	mux     sync.Mutex
	least   float64
	most    float64
	entries []*LeaderboardEntry
}

// LeaderboardEntry is sorted in Leaderboard entries by the target.
type LeaderboardEntry struct {
	name   string
	target float64
	extra  map[string]interface{}
}

// New creates a new leaderboard.
func New(name string, maxEntries int, dupePushAbove bool, alwaysReplace bool) (*Leaderboard, int) {
	leaderboardsMux.Lock()
	if leaderboards[name] != nil {
		leaderboardsMux.Unlock()
		return nil, helpers.ErrorLeaderboardExists
	}
	lb := &Leaderboard{maxEntries: maxEntries, dupePushAbove: dupePushAbove, alwaysReplace: alwaysReplace, entries: make([]*LeaderboardEntry, 0)}
	leaderboards[name] = lb
	leaderboardsMux.Unlock()

	return lb, 0
}

// Get retrieves a leaderboard by name.
func Get(name string) (*Leaderboard, int) {
	leaderboardsMux.Lock()
	if leaderboards[name] == nil {
		leaderboardsMux.Unlock()
		return nil, helpers.ErrorLeaderboardDoesntExist
	}
	lb := leaderboards[name]
	leaderboardsMux.Unlock()
	return lb, 0
}

// Len returns the length of the leaderboard
func (l *Leaderboard) Len() int {
	l.mux.Lock()
	length := len(l.entries)
	l.mux.Unlock()
	return length
}

// GetPage gets a segment of the list for paging.
func (l *Leaderboard) GetPage(limit int, page int) []LeaderboardEntry {
	if page < 0 {
		page = 0
	}
	var p []*LeaderboardEntry
	// Get entries
	l.mux.Lock()
	if page*limit >= len(l.entries) {
		p = make([]*LeaderboardEntry, 0)
	} else if (page+1)*limit > len(l.entries) {
		p = l.entries[page*limit:]
	} else {
		p = l.entries[page*limit : (page+1)*limit]
	}
	l.mux.Unlock()
	// Convert to non-pointer list
	cp := make([]LeaderboardEntry, len(p), len(p))
	for i := 0; i < len(p); i++ {
		cp[i] = *p[i]
	}
	return cp
}

// CheckAndPush checks the target against the leaderboard, and pushes worth entries into it.
func (l *Leaderboard) CheckAndPush(name string, target float64, extra map[string]interface{}) {
	l.mux.Lock()
	// Check if Leaderboard is empty
	entriesLen := len(l.entries)
	if entriesLen == 0 {
		newEntry := LeaderboardEntry{name: name, target: target, extra: extra}
		l.entries = []*LeaderboardEntry{&newEntry}
		l.least = target
		l.most = target
		l.mux.Unlock()
		return
	}
	// Check if entry is worthy and for previous entry by same name
	previousPos := -1
	var previousTarget float64 = -1
	newPos := -1
	if target > l.most {
		newPos = 0
	} else if entriesLen < l.maxEntries && target < l.least {
		newPos = entriesLen
	} else if entriesLen == l.maxEntries && target < l.least {
		newPos = -2
	}
	for i := 0; i < entriesLen; i++ {
		currTarget := l.entries[i].target
		if l.entries[i].name == name {
			previousPos = i
			previousTarget = currTarget
		}
		if newPos == -1 {
			if target > currTarget {
				newPos = i
			} else if target == currTarget {
				if l.dupePushAbove {
					newPos = i
				} else {
					if i == l.maxEntries-1 {
						newPos = -2
					} else if i == entriesLen-1 {
						newPos = i + 1
					} else {
						for j := i + 1; j < entriesLen; j++ {
							if target > l.entries[j].target {
								newPos = j
								break
							} else if j == entriesLen-1 {
								if entriesLen < l.maxEntries {
									newPos = j + 1
								} else {
									newPos = -2
								}
							}
						}
					}
				}
			}
		}
		if newPos != -1 && previousPos != -1 {
			break
		}
	}
	// Apply any changes
	if newPos >= 0 {
		newEntry := LeaderboardEntry{name: name, target: target, extra: extra}
		if previousPos >= 0 {
			if previousPos > newPos || (previousPos < newPos && l.alwaysReplace) {
				// move previousPos to newPos
				if previousPos < newPos {
					newPos--
				}
				l.entries = append(l.entries[:previousPos], l.entries[previousPos+1:]...)
				l.entries = append(l.entries[:newPos], append([]*LeaderboardEntry{&newEntry}, l.entries[newPos:]...)...)
				l.least = l.entries[len(l.entries)-1].target
				l.most = l.entries[0].target
			} else if previousPos == newPos && (previousTarget < target || l.alwaysReplace) {
				// replace previousPos
				l.entries[previousPos] = &newEntry
				l.least = l.entries[len(l.entries)-1].target
				l.most = l.entries[0].target
			}
		} else {
			// insert to newPos
			l.entries = append(l.entries[:newPos], append([]*LeaderboardEntry{&newEntry}, l.entries[newPos:]...)...)
			//remove last item if too large
			if len(l.entries) > l.maxEntries {
				l.entries = l.entries[:l.maxEntries]
			}
			l.least = l.entries[len(l.entries)-1].target
			l.most = l.entries[0].target
		}
	} else if previousPos >= 0 && l.alwaysReplace {
		// move previousPos to end
		newEntry := LeaderboardEntry{name: name, target: target, extra: extra}
		l.entries = append(l.entries[:previousPos], l.entries[previousPos+1:]...)
		l.entries = append(l.entries, &newEntry)
		l.least = target
	}
	l.mux.Unlock()
}

// Print prints the leaderboard to console.
func (l *Leaderboard) Print() {
	fmt.Println("=================================================")
	l.mux.Lock()
	fmt.Println("[ Least:", l.least, "]")
	for i := 0; i < len(l.entries); i++ {
		fmt.Println(l.entries[i].name, "|", l.entries[i].target, "|", l.entries[i].extra)
	}
	l.mux.Unlock()
	fmt.Println("=================================================")
}
