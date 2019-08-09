package gossip

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

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

	for i := 0; i < 100; i++ {
		nodeIndex := rand.Uint32() % 100
		nodes[nodeIndex].Gossip([]byte(strconv.Itoa(i)))
	}

	fmt.Println("messages sent")

	for i := range nodes {
		var check map[string]bool
		for j := 0; j < 100; j++ {
			msg := string(nodes[i].GetMsg())
			if _, ok := check[msg]; ok {
				t.Error("duplicated message")
			}
		}
		fmt.Printf("node %d complete\n", i)
	}
}
