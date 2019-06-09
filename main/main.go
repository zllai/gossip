package main

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/zllai/gossip"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage:\n\tgossip [topic] [listenIP:Port] [BootNodeIP:Port]...")
		os.Exit(1)
	}
	topic := os.Args[1]
	listenAddr, err := net.ResolveUDPAddr("udp", os.Args[2])
	if err != nil {
		fmt.Printf("Cannot resolve listen address: %s\n", err.Error())
		os.Exit(1)
	}
	bootNodes := make([]net.Addr, len(os.Args)-3)
	for i := 3; i < len(os.Args); i++ {
		bootNodes[i-3], err = net.ResolveUDPAddr("udp", os.Args[i])
		if err != nil {
			fmt.Printf("Cannot resolve bootnode address %s: %s\n", os.Args[i], err.Error())
			os.Exit(1)
		}
	}
	node := gossip.New(listenAddr, topic)
	node.Listen()
	node.Join(bootNodes)
	node.StartDiscover()
	reader := bufio.NewReader(os.Stdin)
	go func() {
		for {
			msg := node.GetMsg()
			msg = msg[:len(msg)-4]
			fmt.Println(string(msg))
		}
	}()

	input, _, err := reader.ReadLine()
	for err == nil {
		nonce := make([]byte, 4)
		rand.Read(nonce)
		err = node.Gossip(append(input, nonce...))
		if err != nil {
			log.Println(err.Error())
		}
		input, _, err = reader.ReadLine()
	}

}
