package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/zllai/gossip"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage:\n\tgossip [topic] [listenIP:Port] [existingNodeIP:Port]...")
		os.Exit(1)
	}
	topic := os.Args[1]
	addr := os.Args[2]
	bootNodes := make([]gossip.NodeId, len(os.Args)-3)
	for i := range bootNodes {
		bootNodes[i] = gossip.NewNodeId(os.Args[i+3])
	}
	node := gossip.New(gossip.NewNodeId(addr), topic)
	node.Join(bootNodes)
	go node.Listen()
	reader := bufio.NewReader(os.Stdin)
	go func() {
		for {
			msg := node.GetMsg()
			fmt.Println(string(msg))
		}
	}()

	input, _, err := reader.ReadLine()
	for err == nil {
		node.Gossip(input)
		input, _, err = reader.ReadLine()
	}

}
