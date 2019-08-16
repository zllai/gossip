package gossip

import (
	"container/list"
	"errors"
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
	connPool  map[NodeId]*grpc.ClientConn
	blackList *set.Set
	lock      *sync.RWMutex
}

func NewNeighborList(cap int) *NeighborList {
	return &NeighborList{
		cap:       cap,
		neighbors: list.New(),
		connPool:  make(map[NodeId]*grpc.ClientConn),
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

	if _, ok := nl.connPool[nodeId]; !ok {
		conn, err := nodeId.Dial()
		if err != nil {
			log.Printf("[gossip] cannot dial node %s: %s", nodeId.String(), err.Error())
			return
		}
		nl.neighbors.PushFront(nodeId)
		nl.connPool[nodeId] = conn
		if nl.neighbors.Len() > nl.cap {
			last := nl.neighbors.Back()
			nl.connPool[last.Value.(NodeId)].Close()
			delete(nl.connPool, last.Value.(NodeId))
			nl.neighbors.Remove(last)
		}
	}

	var e *list.Element
	for e = nl.neighbors.Front(); e != nil; e = e.Next() {
		if e.Value.(NodeId) == nodeId {
			nl.neighbors.MoveToFront(e)
			break
		}
	}
}

func (nl *NeighborList) Len() int {
	return nl.neighbors.Len()
}

func (nl *NeighborList) SampleIdString(num int) []string {
	nl.lock.RLock()
	defer nl.lock.RUnlock()
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
	nl.lock.RLock()
	defer nl.lock.RUnlock()
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

func (nl *NeighborList) Reconnect(nodeId NodeId) {
	nl.lock.Lock()
	defer nl.lock.Unlock()
	for e := nl.neighbors.Front(); e != nil; e = e.Next() {
		if e.Value.(NodeId) == nodeId {
			nl.connPool[nodeId].Close()
			var err error
			nl.connPool[nodeId], err = nodeId.Dial()
			if err != nil {
				log.Printf("[gossip] Cannot dial node %s: %s", nodeId.String(), err.Error())
				nl.neighbors.Remove(e)
				delete(nl.connPool, nodeId)
				return
			}
			nl.neighbors.MoveToBack(e)
			break
		}
	}
}

func (nl *NeighborList) Print() {
	nl.lock.RLock()
	defer nl.lock.RUnlock()
	for e := nl.neighbors.Front(); e != nil; e = e.Next() {
		fmt.Printf("%s\n", e.Value.(NodeId).String())
	}
}

func (nl *NeighborList) GetNeighborsId() []NodeId {
	nl.lock.RLock()
	defer nl.lock.RUnlock()
	ret := make([]NodeId, 0)
	for e := nl.neighbors.Front(); e != nil; e = e.Next() {
		ret = append(ret, e.Value.(NodeId))
	}
	return ret
}

func (nl *NeighborList) GetConn(nodeId NodeId) (*grpc.ClientConn, error) {
	nl.lock.RLock()
	defer nl.lock.RUnlock()
	if conn, ok := nl.connPool[nodeId]; ok {
		return conn, nil
	} else {
		return nil, errors.New("conn not found")
	}
}
