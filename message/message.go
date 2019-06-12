package message

import (
	fmt "fmt"
	"net"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
)

func ResAddr2Addr(resAddr *NeighborRes_Addr) net.Addr {
	addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", resAddr.Ip, resAddr.Port))
	return addr
}

func Addr2resAddr(addr net.Addr) *NeighborRes_Addr {
	sep := strings.LastIndexByte(addr.String(), ':')
	Ip := addr.String()[:sep]
	Port, _ := strconv.Atoi(addr.String()[sep+1:])
	return &NeighborRes_Addr{
		Ip:   Ip,
		Port: int32(Port),
	}
}

func ResponseNeighborList(conn net.PacketConn, addr net.Addr, topic string, samples []net.Addr) error {
	neighborRes := NeighborRes{
		Nodes: make([]*NeighborRes_Addr, len(samples)),
	}
	for i := 0; i < len(samples); i++ {
		neighborRes.Nodes[i] = Addr2resAddr(samples[i])
	}
	res := &GossipMsg{
		Topic:   topic,
		Content: &GossipMsg_NeighborRes{NeighborRes: &neighborRes},
	}
	resData, err := proto.Marshal(res)
	if err != nil {
		return err
	}
	conn.WriteTo(resData, addr)
	return nil
}

func RequestNeighborList(conn net.PacketConn, addr net.Addr, topic string, maxNum int) error {
	req := &GossipMsg{
		Topic:   topic,
		Content: &GossipMsg_NeighborReq{NeighborReq: &NeighborReq{MaxNum: int32(maxNum)}},
	}
	data, err := proto.Marshal(req)
	if err != nil {
		return err
	}
	conn.WriteTo(data, addr)
	return nil
}

func GossipToNodes(conn net.PacketConn, addrs []net.Addr, topic string, data []byte, nonce []byte) error {
	pkt := &GossipMsg{
		Topic: topic,
		Content: &GossipMsg_Data{Data: &Data{
			Nonce:   nonce,
			Payload: data,
		}},
	}
	pktBytes, err := proto.Marshal(pkt)
	if err != nil {
		return err
	}
	for i := 0; i < len(addrs); i++ {
		conn.WriteTo(pktBytes, addrs[i])
	}
	return nil
}
