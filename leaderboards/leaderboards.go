package leaderboards

import (
	"sync"
	"github.com/hewiefreeman/GopherGameDB/helpers"
	"fmt"
)

var (
	leaderboardsMux sync.Mutex
	leaderboards map[string]*Leaderboard = make(map[string]*Leaderboard)
)

type Leaderboard struct {
	maxEntries    int
	dupePush      int
	alwaysReplace bool

	mux     sync.Mutex
	least   int
	entries []*LeaderboardEntry
}

// LeaderboardEntry is sorted in Leaderboard entries by the target.
type LeaderboardEntry struct {
	name   string
	target int
	extra  map[string]interface{}
}

const (
	DuplicateTargetPushTop = iota
	DuplicateTargetPushBottom
)

func New(name string, maxEntries int, dupePush int, alwaysReplace bool) (*Leaderboard, int) {
	dt := 0
	if dupePush > 0 {
		dt = 1
	}
	leaderboardsMux.Lock()
	if leaderboards[name] != nil {
		leaderboardsMux.Unlock()
		return nil, helpers.ErrorLeaderboardExists
	}
	lb := &Leaderboard{maxEntries: maxEntries, dupePush: dt, alwaysReplace: alwaysReplace, entries: make([]*LeaderboardEntry, 0)}
	leaderboards[name] = lb
	leaderboardsMux.Unlock()

	return lb, 0
}

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

func (l *Leaderboard) Len() int {
	l.mux.Lock()
	length := len(l.entries)
	l.mux.Unlock()
	return length
}

func (l *Leaderboard) CheckAndPush(name string, target int, extra map[string]interface{}) {
	l.mux.Lock()
	// Check if Leaderboard is empty
	entriesLen := len(l.entries)
	if entriesLen == 0 {
		newEntry := LeaderboardEntry{name: name, target: target, extra: extra}
		l.entries = []*LeaderboardEntry{&newEntry}
		l.least = target
		l.mux.Unlock()
		return
	}
	// Check if entry is worthy and for previous entry by same name
	previousPos := -1
	previousTarget := -1
	newPos := -1
	if entriesLen < l.maxEntries && target < l.least {
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
				if l.dupePush == DuplicateTargetPushTop {
					newPos = i
				} else {
					if i == l.maxEntries-1 {
						newPos = -2
					} else if i == entriesLen-1 {
						newPos = i+1
					} else {
						for j := i+1; j < entriesLen; j++ {
							if target > l.entries[j].target {
								newPos = j
								break
							} else if j == entriesLen-1 {
								newPos = -2
							}
						}
						fmt.Println(newPos)
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
			} else if previousPos == newPos && (previousTarget < target || l.alwaysReplace) {
				// replace previousPos
				l.entries[previousPos] = &newEntry
				l.least = l.entries[len(l.entries)-1].target
			}
		} else {
			// insert to newPos
			l.entries = append(l.entries[:newPos], append([]*LeaderboardEntry{&newEntry}, l.entries[newPos:]...)...)
			//remove last item if too large
			if len(l.entries) > l.maxEntries {
				l.entries = l.entries[:l.maxEntries]
			}
			l.least = l.entries[len(l.entries)-1].target
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

func (l *Leaderboard) Print() {
	fmt.Println("=================================================")
	l.mux.Lock()
	fmt.Println("[ Least:", l.least, "]")
	for i := 0; i < len(l.entries); i++ {
		fmt.Println(l.entries[i].name, "|", l.entries[i].target)
	}
	l.mux.Unlock()
	fmt.Println("=================================================")
}
