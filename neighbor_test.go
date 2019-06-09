package gossip

import (
	fmt "fmt"
	"net"
	"testing"
	"time"
)

func TestUpdate(t *testing.T) {
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:10000")
	l := NewNeighborList(3, addr)
	addr1, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1000")
	addr2, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1001")
	addr3, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1002")
	addr4, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1000")
	addr5, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1004")
	l.Update(addr1)
	time.Sleep(1e9)
	l.Update(addr2)
	time.Sleep(1e9)
	l.Update(addr3)
	time.Sleep(1e9)
	l.Update(addr4)
	time.Sleep(1e9)
	l.Update(addr5)
	time.Sleep(1e9)
	l.Print()
	if _, ok := l.ActiveRecord[addr2.String()]; ok {
		t.Error("addr2 should be deleted")
	}
}

func TestSample(t *testing.T) {
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:10000")
	l := NewNeighborList(10, addr)
	addr1, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1000")
	addr2, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1001")
	addr3, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1002")
	addr4, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1003")
	addr5, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1004")
	addr6, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1005")
	addr7, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1006")
	l.Update(addr1)
	l.Update(addr2)
	l.Update(addr3)
	l.Update(addr4)
	l.Update(addr5)
	l.Update(addr6)
	l.Update(addr7)
	l.Print()
	sample := l.Sample(3)
	fmt.Println("samples:")
	for i := 0; i < len(sample); i++ {
		fmt.Println(sample[i].String())
	}
	sample = l.Sample(8)
	if len(sample) != 7 {
		t.Errorf("sample size incorrect, size is %d", len(sample))
	}
}
