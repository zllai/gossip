package gossip

import (
	context "context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

const bufferCap = 256
const neighborListCap = 256
const gossipFanout = 16
const discoveryFanout = 8
const broadcastFanout = neighborListCap

type NodeId string

func NewNodeId(str string) NodeId {
	return NodeId(str)
}

func (id NodeId) Dial() (*grpc.ClientConn, error) {
	return grpc.Dial(string(id), grpc.WithInsecure())
}

func (id NodeId) String() string {
	return string(id)
}

type Node struct {
	topic     string
	nodeId    NodeId
	neighbors *NeighborList
	msgFilter *Filter
	msgChan   chan []byte
}

func New(nodeId NodeId, topic string) *Node {
	node := &Node{
		topic:     topic,
		nodeId:    nodeId,
		neighbors: NewNeighborList(neighborListCap),
		msgChan:   make(chan []byte, bufferCap),
		msgFilter: NewFilter(60),
	}
	node.neighbors.AddBlackList(nodeId)
	return node
}

func (node *Node) Listen() error {
	lis, err := net.Listen("tcp", node.nodeId.String())
	if err != nil {
		return errors.New(fmt.Sprintf("[gossip] Cannot listen on %s: %s", node.nodeId.String(), err.Error()))
	}
	grpcServer := grpc.NewServer()
	RegisterGossipServer(grpcServer, node)
	grpcServer.Serve(lis)
	return nil
}

func (node *Node) Register(grpcServer *grpc.Server) {
	RegisterGossipServer(grpcServer, node)
}

func (node *Node) GetPeers(ctx context.Context, req *NeighborReq) (*NeighborRes, error) {
	if req.Topic != node.topic {
		return nil, status.Errorf(codes.NotFound, "[From %s] topic does not match", node.nodeId.String())
	}
	nodeId := NewNodeId(req.NodeId)
	node.neighbors.Update(nodeId)
	samples := node.neighbors.SampleIdString(int(req.MaxNum))
	res := &NeighborRes{
		Topic:     node.topic,
		NodeId:    node.nodeId.String(),
		Neighbors: samples,
	}
	return res, nil
}

func (node *Node) SendData(ctx context.Context, data *GossipData) (*Empty, error) {
	if data.Topic != node.topic {
		return nil, status.Errorf(codes.NotFound, "[From %s] topic does not match", node.nodeId.String())
	}
	nodeId := NewNodeId(data.NodeId)
	node.neighbors.Update(nodeId)

	// check redundancy and store in buffer
	if !node.msgFilter.Check(data.Hash()) {
		return nil, status.Errorf(codes.NotFound, "[From %s] already received the same message", node.nodeId.String())
	}
	node.msgChan <- data.Payload

	//gossip to other nodes
	node.gossipToPeers(data, gossipFanout)
	return &Empty{}, nil
}

func (node *Node) gossipToPeers(data *GossipData, fanout int) {
	nodeIds := node.neighbors.SampleNodeId(fanout)
	for i := range nodeIds {
		go func(nodeId NodeId) {
			conn, err := node.neighbors.GetConn(nodeId)
			if err != nil {
				log.Printf("[gossip] Connection to %s is closed", nodeId.String())
				return
			}
			client := NewGossipClient(conn)
			_, err = client.SendData(context.Background(), data)
			if err != nil && status.Convert(err).Code() != codes.NotFound {
				log.Printf("[gossip] Cannot send data to node %s: %s", nodeId.String(), err.Error())
				node.neighbors.Reconnect(nodeId)
			}
		}(nodeIds[i])
	}
}

func (node *Node) Join(bootnodes []NodeId) error {
	// add to neighbor list
	for i := range bootnodes {
		node.neighbors.Update(bootnodes[i])
	}

	go func() {
		// run discovery forever
		for {
			// whether not enough peers
			if node.neighbors.Len() >= neighborListCap {
				time.Sleep(5 * time.Second)
				continue
			}

			// how many peers to ask
			fanout := discoveryFanout
			if fanout > node.neighbors.Len() {
				fanout = node.neighbors.Len()
			}

			// construct request
			avgReqests := int(float32(neighborListCap-node.neighbors.Len()) / float32(fanout) * 1.2)
			req := &NeighborReq{
				Topic:  node.topic,
				NodeId: node.nodeId.String(),
				MaxNum: int32(avgReqests),
			}

			nodeIds := node.neighbors.SampleNodeId(fanout)
			for i := range nodeIds {
				go func(nodeId NodeId) {
					conn, err := node.neighbors.GetConn(nodeId)
					if err != nil {
						log.Printf("[gossip] connection to %s is closed", nodeId.String())
						return
					}
					client := NewGossipClient(conn)
					res, err := client.GetPeers(context.Background(), req)
					if err != nil {
						log.Printf("[gossip] node %s cannot call GetPeer: %s", nodeId.String(), err.Error())
						node.neighbors.Reconnect(nodeId)
						return
					}
					for j := range res.Neighbors {
						node.neighbors.Update(NewNodeId(res.Neighbors[j]))
					}
				}(nodeIds[i])
			}
			time.Sleep(5 * time.Second)
		}
	}()

	return nil
}

func (node *Node) Gossip(data []byte) {
	nonce := rand.Uint64()
	gossipData := &GossipData{
		Topic:   node.topic,
		NodeId:  node.nodeId.String(),
		Nonce:   nonce,
		Payload: data,
	}

	// gossip to self
	if node.msgFilter.Check(gossipData.Hash()) {
		node.msgChan <- data
	}

	node.gossipToPeers(gossipData, broadcastFanout)
}

func (node *Node) GetMsgChan() chan []byte {
	return node.msgChan
}

func (node *Node) PrintPeers() {
	node.neighbors.Print()
}

func (node *Node) GetNeighborList(NodeId) *NeighborList {
	return node.neighbors
}
