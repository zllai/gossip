package gossip

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/zllai/gossip/filter"
	"github.com/zllai/gossip/message"
	"github.com/zllai/gossip/ringbuffer"
)

const bufferCap = 5242880 // 5MB
const neighborListCap = 256
const gossipFanout = 64
const broadcastFanout = 256

type Node struct {
	topic       string
	addr        net.Addr
	conn        net.PacketConn
	neighbors   *NeighborList
	discovering bool
	msgBuffer   *ringbuffer.RingBuffer
	msgFilter   *filter.Filter
}

func New(addr net.Addr, topic string) *Node {
	conn, err := net.ListenPacket("udp", addr.String())
	if err != nil {
		log.Fatalf("Cannot create udp listener: %s", err.Error())
	}
	return &Node{
		topic:       topic,
		addr:        addr,
		conn:        conn,
		neighbors:   NewNeighborList(neighborListCap, addr),
		discovering: false,
		msgBuffer:   ringbuffer.New(64),
		msgFilter:   filter.New(60),
	}
}

func (node *Node) serve(addr net.Addr, data []byte) error {
	//Deserialize data
	msg := message.GossipMsg{}
	err := proto.Unmarshal(data, &msg)
	if err != nil {
		return err
	}
	if msg.Topic != node.topic {
		return errors.New("Topic mismatch")
	}
	switch content := msg.Content.(type) {
	case *message.GossipMsg_NeighborReq:
		samples := node.neighbors.Sample(int(content.NeighborReq.MaxNum))
		message.ResponseNeighborList(node.conn, addr, node.topic, samples)
		if node.neighbors.Update(addr) {
			log.Printf("New node : %s", addr.String())
		}
	case *message.GossipMsg_NeighborRes:
		if node.neighbors.Len() < node.neighbors.Capacity {
			neighborRes := content.NeighborRes.Nodes
			for i := 0; i < len(neighborRes); i++ {
				newNeighbor := message.ResAddr2Addr(neighborRes[i])
				if node.neighbors.Update(newNeighbor) {
					log.Printf("New node : %s", newNeighbor.String())
				}
			}
			if node.neighbors.Update(addr) {
				log.Printf("New node : %s", addr.String())
			}
		}
	case *message.GossipMsg_Data:
		data := content.Data.Payload
		nonce := content.Data.Nonce
		md5Func := md5.New()
		hash := md5Func.Sum(append(data, nonce...))
		if node.msgFilter.Check(hex.EncodeToString(hash)) {
			err = message.GossipToNodes(node.conn, node.neighbors.Sample(gossipFanout), node.topic, data, nonce)
			if err != nil {
				log.Printf("Cannot gossip message: %s", err.Error())
			}
			node.msgBuffer.Push(content.Data.Payload)
		}
	}

	return nil
}

func (node *Node) Listen() {
	go func() {
		buf := make([]byte, bufferCap)
		for {
			size, addr, err := node.conn.ReadFrom(buf)
			if err != nil {
				log.Printf("Discard messge: %s", err.Error())
				continue
			}
			if size == bufferCap {
				log.Printf("Discard message: message larger than buffer size")
				continue
			}
			err = node.serve(addr, buf[:size])
			if err != nil {
				log.Printf("Invalid message: %s", err.Error())
			}
		}
	}()
}

func (node *Node) Close() {
	node.conn.Close()
}

func (node *Node) Join(bootnodes []net.Addr) error {
	avgReqests := int(float32(neighborListCap) / float32(len(bootnodes)) * 1.2)
	for i := 0; i < len(bootnodes); i++ {
		err := message.RequestNeighborList(node.conn, bootnodes[i], node.topic, avgReqests)
		if err != nil {
			log.Printf("Cannot connect to bootnode %s: %s", bootnodes[i].String(), err.Error())
		}
	}
	return nil
}

func (node *Node) StartDiscover() error {
	if node.discovering {
		return errors.New("Peer discovery already started")
	}
	node.discovering = true
	go func() {
		for node.discovering {
			if node.neighbors.Len() >= neighborListCap {
				continue
			}
			samples := node.neighbors.Sample(5)
			for i := 0; i < len(samples); i++ {
				err := message.RequestNeighborList(node.conn, samples[i], node.topic, 10)
				if err != nil {
					log.Printf("Cannot connect to peer %s: %s", samples[i].String(), err.Error())
				}
			}
			time.Sleep(10e9)
		}
	}()
	return nil
}

func (node *Node) StopDiscover() {
	node.discovering = false
}

func (node *Node) Gossip(data []byte) error {
	nonce := make([]byte, 4)
	_, err := rand.Read(nonce)
	if err != nil {
		return err
	}
	md5Func := md5.New()
	hash := md5Func.Sum(append(data, nonce...))
	node.msgFilter.Check(hex.EncodeToString(hash))
	err = message.GossipToNodes(node.conn, node.neighbors.Sample(broadcastFanout), node.topic, data, nonce)
	return err
}

func (node *Node) GetMsg() []byte {
	return node.msgBuffer.Pop()
}

func (node *Node) PrintPeers() {
	node.neighbors.Print()
}
