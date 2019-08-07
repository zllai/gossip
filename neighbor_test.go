package gossip

import (
	fmt "fmt"
	"testing"
	"time"
)

func TestUpdate(t *testing.T) {
	l := NewNeighborList(3)
	l.Update(NewNodeId("127.0.0.1:8000"))
	l.Update(NewNodeId("127.0.0.1:8001"))
	l.Update(NewNodeId("127.0.0.1:8002"))
	go l.Update(NewNodeId("127.0.0.1:8003"))
	go l.Update(NewNodeId("127.0.0.1:8003"))
	go l.Update(NewNodeId("127.0.0.1:8003"))
	go l.Update(NewNodeId("127.0.0.1:8002"))
	time.Sleep(time.Second)
	l.Print()
}

func TestSample(t *testing.T) {
	l := NewNeighborList(10)
	l.Update(NewNodeId("127.0.0.1:8001"))
	l.Update(NewNodeId("127.0.0.1:8002"))
	l.Update(NewNodeId("127.0.0.1:8003"))
	l.Update(NewNodeId("127.0.0.1:8004"))
	l.Update(NewNodeId("127.0.0.1:8005"))
	l.Update(NewNodeId("127.0.0.1:8006"))
	l.Update(NewNodeId("127.0.0.1:8007"))
	l.Print()
	sample := l.SampleIdString(3)
	fmt.Println("samples:")
	for i := 0; i < len(sample); i++ {
		fmt.Println(sample[i])
	}

	sample = l.SampleIdString(8)
	if len(sample) != 7 {
		t.Errorf("sample size incorrect, size is %d", len(sample))
	}
}
