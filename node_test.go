package gossip

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"testing"
	"time"
)

func TestJoin(t *testing.T) {
	addr1, _ := net.ResolveUDPAddr("udp", "127.0.0.1:8001")
	addr2, _ := net.ResolveUDPAddr("udp", "127.0.0.1:8002")
	addr3, _ := net.ResolveUDPAddr("udp", "127.0.0.1:8003")
	addr4, _ := net.ResolveUDPAddr("udp", "127.0.0.1:8004")
	addr5, _ := net.ResolveUDPAddr("udp", "127.0.0.1:8005")
	addr6, _ := net.ResolveUDPAddr("udp", "127.0.0.1:8006")
	node1 := New(addr1, "1.0.0")
	node2 := New(addr2, "1.0.0")
	node3 := New(addr3, "1.0.0")
	node4 := New(addr4, "1.0.0")
	node5 := New(addr5, "1.0.0")
	node6 := New(addr6, "1.0.0")
	node1.Listen()
	node2.Listen()
	node3.Listen()
	node4.Listen()
	node5.Listen()
	node6.Listen()
	node2.Join([]net.Addr{addr1})
	time.Sleep(1e8)
	node3.Join([]net.Addr{addr1})
	time.Sleep(1e8)
	node4.Join([]net.Addr{addr1})
	time.Sleep(1e8)
	node5.Join([]net.Addr{addr1})
	time.Sleep(1e8)
	node6.Join([]net.Addr{addr1})
	time.Sleep(1e8)
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

func TestDiscover(t *testing.T) {
	addr1, _ := net.ResolveUDPAddr("udp", "127.0.0.1:8001")
	addr2, _ := net.ResolveUDPAddr("udp", "127.0.0.1:8002")
	addr3, _ := net.ResolveUDPAddr("udp", "127.0.0.1:8003")
	addr4, _ := net.ResolveUDPAddr("udp", "127.0.0.1:8004")
	addr5, _ := net.ResolveUDPAddr("udp", "127.0.0.1:8005")
	addr6, _ := net.ResolveUDPAddr("udp", "127.0.0.1:8006")
	node1 := New(addr1, "1.0.0")
	node2 := New(addr2, "1.0.0")
	node3 := New(addr3, "1.0.0")
	node4 := New(addr4, "1.0.0")
	node5 := New(addr5, "1.0.0")
	node6 := New(addr6, "1.0.0")
	node1.Listen()
	node2.Listen()
	node3.Listen()
	node4.Listen()
	node5.Listen()
	node6.Listen()
	node2.Join([]net.Addr{addr1})
	time.Sleep(1e8)
	node3.Join([]net.Addr{addr1})
	time.Sleep(1e8)
	node4.Join([]net.Addr{addr1})
	time.Sleep(1e8)
	node5.Join([]net.Addr{addr1})
	time.Sleep(1e8)
	node6.Join([]net.Addr{addr1})
	time.Sleep(1e8)
	node2.StartDiscover()
	node3.StartDiscover()
	node4.StartDiscover()
	node5.StartDiscover()
	node6.StartDiscover()
	time.Sleep(1e9)
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
	bootAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:7999")
	bootNode := New(bootAddr, "test topic")
	bootNode.Listen()
	var nodes [100]*Node
	for i := range nodes {
		addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", 8000+i))
		nodes[i] = New(addr, "test topic")
		nodes[i].Join([]net.Addr{bootAddr})
		nodes[i].Listen()
		nodes[i].StartDiscover()
	}
	time.Sleep(20 * time.Second)
	fmt.Printf("bootNode neighbors: %d\n", bootNode.neighbors.Len())
	for i := range nodes {
		fmt.Printf("node %d neighbors: %d\n", i, nodes[i].neighbors.Len())
	}
	for i := 0; i < 30; i++ {
		nodeIndex := rand.Uint32() % 100
		nodes[nodeIndex].Gossip([]byte(strconv.Itoa(i)))
	}
	fmt.Println("messages sent")
	time.Sleep(2 * time.Second)
	for i := range nodes {
		var check map[string]bool
		for j := 0; j < 30; j++ {
			msg := string(nodes[i].GetMsg())
			if _, ok := check[msg]; ok {
				t.Error("duplicated message")
			}
		}
		fmt.Printf("node %d complete\n", i)
	}
}
