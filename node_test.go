package gossip

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestJoin(t *testing.T) {
	bootNode := NewNodeId("127.0.0.1:8001")
	node1 := New(bootNode, "1.0.0")
	node2 := New(NewNodeId("127.0.0.1:8002"), "1.0.0")
	node3 := New(NewNodeId("127.0.0.1:8003"), "1.0.0")
	node4 := New(NewNodeId("127.0.0.1:8004"), "1.0.0")
	node5 := New(NewNodeId("127.0.0.1:8005"), "1.0.0")
	node6 := New(NewNodeId("127.0.0.1:8006"), "1.0.0")

	go node1.Listen()
	go node2.Listen()
	go node3.Listen()
	go node4.Listen()
	go node5.Listen()
	go node6.Listen()

	node2.Join([]NodeId{bootNode})
	node3.Join([]NodeId{bootNode})
	node4.Join([]NodeId{bootNode})
	node5.Join([]NodeId{bootNode})
	node6.Join([]NodeId{bootNode})
	time.Sleep(10 * time.Second)
	fmt.Println("neighbors of node 1:")
	node1.PrintPeers()
	fmt.Println("neighbors of node 2:")
	node2.PrintPeers()
	fmt.Println("neighbors of node 3:")
	node3.PrintPeers()
	fmt.Println("neighbors of node 4:")
	node4.PrintPeers()
	fmt.Println("neighbors of node 5:")
	node5.PrintPeers()
	fmt.Println("neighbors of node 6:")
	node6.PrintPeers()

	time.Sleep(10 * time.Second)
	fmt.Println("After 10 seconds")
	fmt.Println("neighbors of node 1:")
	node1.PrintPeers()
	fmt.Println("neighbors of node 2:")
	node2.PrintPeers()
	fmt.Println("neighbors of node 3:")
	node3.PrintPeers()
	fmt.Println("neighbors of node 4:")
	node4.PrintPeers()
	fmt.Println("neighbors of node 5:")
	node5.PrintPeers()
	fmt.Println("neighbors of node 6:")
	node6.PrintPeers()
}

func TestMessages(t *testing.T) {
	bootNodeId := NewNodeId("127.0.0.1:7999")
	bootNode := New(bootNodeId, "test topic")
	go bootNode.Listen()
	var nodes [100]*Node
	for i := range nodes {
		nodeId := NewNodeId(fmt.Sprintf("127.0.0.1:%d", 8000+i))
		nodes[i] = New(nodeId, "test topic")
		go nodes[i].Listen()
		nodes[i].Join([]NodeId{bootNodeId})
	}

	time.Sleep(10 * time.Second)
	fmt.Printf("bootNode neighbors: %d\n", bootNode.neighbors.Len())
	for i := range nodes {
		fmt.Printf("node %d neighbors: %d\n", i, nodes[i].neighbors.Len())
	}

	for i := 0; i < 10; i++ {
		nodeIndex := rand.Uint32() % 100
		nodes[nodeIndex].Gossip([]byte(strconv.Itoa(i)))
	}

	fmt.Println("messages sent")

	for i := range nodes {
		var check map[string]bool
		for j := 0; j < 10; j++ {
			msg := string(nodes[i].GetMsg())
			if _, ok := check[msg]; ok {
				t.Error("duplicated message")
			}
		}
		fmt.Printf("node %d complete\n", i)
	}
}
