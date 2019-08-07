package gossip

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
)

func UInt64ToBytes(a uint64) []byte {
	ret := make([]byte, 8)
	binary.BigEndian.PutUint64(ret, a)
	return ret
}
func (gossipData *GossipData) Hash() string {
	data := bytes.Join([][]byte{
		UInt64ToBytes(gossipData.Nonce),
		gossipData.Payload,
	}, nil)
	hashBytes := md5.Sum(data)
	return hex.EncodeToString(hashBytes[:])
}
