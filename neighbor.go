package gossip

import (
	"container/list"
	"fmt"
	"math/rand"
	"sort"
	"sync"

	"github.com/golang-collections/collections/set"
)

type NeighborList struct {
	cap       int
	neighbors *list.List
	blackList *set.Set
	lock      *sync.RWMutex
}

func NewNeighborList(cap int) *NeighborList {
	return &NeighborList{
		cap:       cap,
		neighbors: list.New(),
		blackList: set.New(),
		lock:      &sync.RWMutex{},
	}
}

func (nl *NeighborList) AddBlackList(nodeId NodeId) {
	nl.lock.Lock()
	defer nl.lock.Unlock()
	nl.blackList.Insert(nodeId)
}

func (nl *NeighborList) Update(nodeId NodeId) {
	nl.lock.Lock()
	defer nl.lock.Unlock()
	if nl.blackList.Has(nodeId) {
		return
	}

	var e *list.Element
	for e = nl.neighbors.Front(); e != nil; e = e.Next() {
		if e.Value.(NodeId) == nodeId {
			nl.neighbors.MoveToFront(e)
			break
		}
	}

	if e == nil {
		nl.neighbors.PushFront(nodeId)
		if nl.neighbors.Len() > nl.cap {
			nl.neighbors.Remove(nl.neighbors.Back())
		}
	}
}

func (nl *NeighborList) Len() int {
	return nl.neighbors.Len()
}

func (nl *NeighborList) SampleIdString(num int) []string {
	nl.lock.Lock()
	defer nl.lock.Unlock()
	len := nl.Len()
	if num > len {
		num = len
	}
	samples := make([]string, num)
	randIndex := rand.Perm(len)[0:num]
	sort.IntSlice(randIndex).Sort()
	i, j := 0, 0
	for e := nl.neighbors.Front(); e != nil; e = e.Next() {
		if i == randIndex[j] {
			samples[j] = e.Value.(NodeId).String()
			j++
			if j >= num {
				break
			}
		}
		i++
	}
	return samples
}

func (nl *NeighborList) SampleNodeId(num int) []NodeId {
	nl.lock.Lock()
	defer nl.lock.Unlock()
	len := nl.Len()
	if num > len {
		num = len
	}
	samples := make([]NodeId, num)
	randIndex := rand.Perm(len)[0:num]
	sort.IntSlice(randIndex).Sort()
	i, j := 0, 0
	for e := nl.neighbors.Front(); e != nil; e = e.Next() {
		if i == randIndex[j] {
			samples[j] = e.Value.(NodeId)
			j++
			if j >= num {
				break
			}
		}
		i++
	}
	return samples
}

func (nl *NeighborList) Delete(nodeId NodeId) {
	nl.lock.Lock()
	defer nl.lock.Unlock()
	for e := nl.neighbors.Front(); e != nil; e = e.Next() {
		if e.Value.(NodeId) == nodeId {
			nl.neighbors.Remove(e)
		}
	}
}

func (nl *NeighborList) Print() {
	nl.lock.Lock()
	defer nl.lock.Unlock()
	for e := nl.neighbors.Front(); e != nil; e = e.Next() {
		fmt.Printf("%s\n", e.Value.(NodeId).String())
	}
}

func (nl *NeighborList) GetNeighborsId() []NodeId {
	nl.lock.Lock()
	defer nl.lock.Unlock()
	ret := make([]NodeId, 0)
	for e := nl.neighbors.Front(); e != nil; e = e.Next() {
		ret = append(ret, e.Value.(NodeId))
	}
	return ret
}
