package gossip

import (
	"fmt"
	"math"
	"math/rand"
	"net"
	"sort"
	"sync"
	"time"
)

type NeighborList struct {
	Capacity     int
	ActiveRecord map[string]int64
	lock         *sync.Mutex
	self         net.Addr
}

type NodeInfo struct {
	Addr       string `json:"addr"`
	LastActive int64  `json:"lastActive"`
}

func NewNeighborList(capacity int, self net.Addr) *NeighborList {
	return &NeighborList{capacity, make(map[string]int64), &sync.Mutex{}, self}
}

func (nl *NeighborList) Update(node net.Addr) bool {
	if node.String() == nl.self.String() {
		return false
	}
	current := time.Now().Unix()
	nl.lock.Lock()
	defer nl.lock.Unlock()
	if _, ok := nl.ActiveRecord[node.String()]; ok {
		nl.ActiveRecord[node.String()] = current
		return false
	} else {
		for len(nl.ActiveRecord) >= nl.Capacity {
			//find least active
			leastActiveTime := int64(math.MaxInt64)
			leastActiveAddr := ""
			for k, v := range nl.ActiveRecord {
				if v < leastActiveTime {
					leastActiveTime = v
					leastActiveAddr = k
				}
			}
			if leastActiveTime != int64(math.MaxInt64) {
				delete(nl.ActiveRecord, leastActiveAddr)
			}
		}
		nl.ActiveRecord[node.String()] = current
		return true
	}
}

func (nl *NeighborList) Len() int {
	return len(nl.ActiveRecord)
}

func (nl *NeighborList) Sample(num int) []net.Addr {
	nl.lock.Lock()
	defer nl.lock.Unlock()
	if num > len(nl.ActiveRecord) {
		num = len(nl.ActiveRecord)
	}
	ret := make([]net.Addr, num)
	randIndex := rand.Perm(len(nl.ActiveRecord))[0:num]
	sort.IntSlice(randIndex).Sort()
	i, j := 0, 0
	for k, _ := range nl.ActiveRecord {
		if i == randIndex[j] {
			ret[j], _ = net.ResolveUDPAddr("udp", k)
			j++
			if j >= num {
				break
			}
		}
		i++
	}
	return ret
}

func (nl *NeighborList) Print() {
	for k, v := range nl.ActiveRecord {
		fmt.Printf("%s\t%d\n", k, v)
	}
}

func (nl *NeighborList) List() []NodeInfo {
	ret := make([]NodeInfo, 0)
	for k, v := range nl.ActiveRecord {
		ret = append(ret, NodeInfo{k, v})
	}
	return ret
}
