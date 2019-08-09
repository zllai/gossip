package gossip

import (
	"container/list"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"sync"

	"github.com/golang-collections/collections/set"
	"google.golang.org/grpc"
)

type NeighborList struct {
	cap       int
	neighbors *list.List
	blackList *set.Set
	lock      *sync.RWMutex
}

type ConnInfo struct {
	nodeId NodeId
	conn   *grpc.ClientConn
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
		if e.Value.(ConnInfo).nodeId == nodeId {
			nl.neighbors.MoveToFront(e)
			break
		}
	}

	if e == nil {
		conn, err := nodeId.Dial()
		if err != nil {
			log.Printf("[gossip] cannot dial node %s: %s", nodeId.String(), err.Error())
			return
		}
		nl.neighbors.PushFront(ConnInfo{nodeId, conn})
		if nl.neighbors.Len() > nl.cap {
			last := nl.neighbors.Back()
			last.Value.(ConnInfo).conn.Close()
			nl.neighbors.Remove(last)
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
			samples[j] = e.Value.(ConnInfo).nodeId.String()
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
			samples[j] = e.Value.(ConnInfo).nodeId
			j++
			if j >= num {
				break
			}
		}
		i++
	}
	return samples
}

func (nl *NeighborList) SampleConnInfo(num int) []ConnInfo {
	nl.lock.Lock()
	defer nl.lock.Unlock()
	len := nl.Len()
	if num > len {
		num = len
	}
	samples := make([]ConnInfo, num)
	randIndex := rand.Perm(len)[0:num]
	sort.IntSlice(randIndex).Sort()
	i, j := 0, 0
	for e := nl.neighbors.Front(); e != nil; e = e.Next() {
		if i == randIndex[j] {
			samples[j] = e.Value.(ConnInfo)
			j++
			if j >= num {
				break
			}
		}
		i++
	}
	return samples
}

func (nl *NeighborList) Reconnect(nodeId NodeId) {
	nl.lock.Lock()
	defer nl.lock.Unlock()
	for e := nl.neighbors.Front(); e != nil; e = e.Next() {
		connInfo := e.Value.(ConnInfo)
		if connInfo.nodeId == nodeId {
			nl.neighbors.Remove(e)
			connInfo.conn.Close()
			var err error
			connInfo.conn, err = connInfo.nodeId.Dial()
			if err != nil {
				log.Printf("[gossip] Cannot dial node %s: %s", nodeId.String(), err.Error())
				return
			}
			nl.neighbors.PushBack(connInfo)
		}
	}
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
		fmt.Printf("%s\n", e.Value.(ConnInfo).nodeId.String())
	}
}

func (nl *NeighborList) GetNeighborsId() []NodeId {
	nl.lock.Lock()
	defer nl.lock.Unlock()
	ret := make([]NodeId, 0)
	for e := nl.neighbors.Front(); e != nil; e = e.Next() {
		ret = append(ret, e.Value.(ConnInfo).nodeId)
	}
	return ret
}
