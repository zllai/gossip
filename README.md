# Gossip broadcast
This is a library implementating gossip protocol for general purpose message broadcast in peer to peer network.

It is similar to the one used in bitcoin or Ethereum blockchain, but this implementation is simpler, cleaner.

## Usage
The usage of this library is straight forward, please see the example

## Example
### build example
```sh
cd example
go build -o node 
```
### run example
usage:
```sh
./node [topic] [listenIP:listenPort] [existingNodeIP:existingNodePort...]
```

Run the first node:
```sh
./node test 127.0.0.1:8000
```

Run second node to join the first node:
```sh
./node test 127.0.0.1:8001 127.0.0.1:8000
```

You can run more nodes and join any existing node with the same topic.

Input messages from stdin, and the messages will be gossiped to all the nodes.
