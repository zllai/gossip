package gossip

import (
	"fmt"
	"sync"
	"time"
)

type Filter struct {
	truncatePeriod int64
	lastTruncate   int64
	msgRecord      map[string]int64
	lock           *sync.Mutex
}

func NewFilter(truncatePeriod int64) *Filter {
	current := time.Now().Unix()
	return &Filter{truncatePeriod, current, make(map[string]int64), &sync.Mutex{}}
}

func (filter *Filter) Check(msgHash string) bool {
	current := time.Now().Unix()
	filter.lock.Lock()
	defer filter.lock.Unlock()
	if recvTime, ok := filter.msgRecord[msgHash]; ok {
		if current-recvTime < filter.truncatePeriod {
			return false
		} else {
			delete(filter.msgRecord, msgHash)
		}
	}
	filter.msgRecord[msgHash] = current
	if current-filter.lastTruncate > filter.truncatePeriod {
		for k, v := range filter.msgRecord {
			if current-v > filter.truncatePeriod {
				delete(filter.msgRecord, k)
			}
		}
		filter.lastTruncate = current
	}
	return true
}

func (filter *Filter) print() {
	for k, v := range filter.msgRecord {
		fmt.Printf("%s\t%d\n", k, v)
	}
}
