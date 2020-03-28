package discv5

import (
	"crypto/ecdsa"
	"encoding/binary"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type simulation struct {
	mu      sync.RWMutex
	nodes   map[NodeID]*Network
	nodectr uint32
}

func NewSimulation() *simulation {
	return &simulation{nodes: make(map[NodeID]*Network)}
}

func (s *simulation) Shutdown() {
	s.mu.RLock()
	alive := make([]*Network, 0, len(s.nodes))
	for _, n := range s.nodes {
		alive = append(alive, n)
	}
	defer s.mu.RUnlock()

	for _, n := range alive {
		n.Close()
	}
}

func (s *simulation) PrintStats() {
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Println("node counter:", s.nodectr)
	fmt.Println("alive nodes:", len(s.nodes))

	// for _, n := range s.nodes {
	// 	fmt.Printf("%x\n", n.tab.self.ID[:8])
	// 	transport := n.conn.(*simTransport)
	// 	fmt.Println("   joined:", transport.joinTime)
	// 	fmt.Println("   sends:", transport.hashctr)
	// 	fmt.Println("   table size:", n.tab.count)
	// }

	/*for _, n := range s.nodes {
		fmt.Println()
		fmt.Printf("*** Node %x\n", n.tab.self.ID[:8])
		n.log.printLogs()
	}*/

}

func (s *simulation) randomNode() *Network {
	s.mu.Lock()
	defer s.mu.Unlock()

	n := rand.Intn(len(s.nodes))
	for _, net := range s.nodes {
		if n == 0 {
			return net
		}
		n--
	}
	return nil
}

func newkey() *ecdsa.PrivateKey {
	key, err := crypto.GenerateKey()
	if err != nil {
		panic("couldn't generate key: " + err.Error())
	}
	return key
}

func (s *simulation) LaunchNode(log bool) *Network {
	var (
		num = s.nodectr
		key = newkey()
		id  = PubkeyID(&key.PublicKey)
		ip  = make(net.IP, 4)
	)
	s.nodectr++
	binary.BigEndian.PutUint32(ip, num)
	ip[0] = 10
	addr := &net.UDPAddr{IP: ip, Port: 30303}

	transport := &simTransport{joinTime: time.Now(), sender: id, senderAddr: addr, sim: s, priv: key}
	net, err := newNetwork(transport, key.PublicKey, "<no database>", nil)
	if err != nil {
		panic("cannot launch new node: " + err.Error())
	}

	s.mu.Lock()
	s.nodes[id] = net
	s.mu.Unlock()

	return net
}

func RandomResolves(s *simulation, net *Network) {
	randtime := func() time.Duration {
		return time.Duration(rand.Intn(5)+2) * time.Second
	}
	lookup := func(target NodeID) bool {
		result := net.Resolve(target)
		return result != nil && result.ID == target
	}

	timer := time.NewTimer(randtime())
	for {
		select {
		case <-timer.C:
			target := s.randomNode().Self().ID
			fmt.Printf("Node %x resolving target %x \n",net.Self().ID[:16],target[:16])
			if !lookup(target) {
				fmt.Printf("node %x: target %x not found \n", net.Self().ID[:16], target[:16])
			}
			timer.Reset(randtime())
		case <-net.closed:
			return
		}
	}
}

func RandomSingleResolve(s *simulation, net *Network) {
	lookup := func(target NodeID) bool {
		result := net.Resolve(target)
		return result != nil && result.ID == target
	}
	target := s.randomNode().Self().ID
	fmt.Printf("Node %x resolving target %x \n",net.Self().ID[:16],target[:16])
	if !lookup(target) {
		fmt.Printf("node %x: target %x not found \n", net.Self().ID[:16], target[:16])
	}

}
type simTransport struct {
	joinTime   time.Time
	sender     NodeID
	senderAddr *net.UDPAddr
	sim        *simulation
	hashctr    uint64
	priv       *ecdsa.PrivateKey
}

func (st *simTransport) localAddr() *net.UDPAddr {
	return st.senderAddr
}

func (st *simTransport) Close() {}

func (st *simTransport) send(remote *Node, ptype nodeEvent, data interface{}) (hash []byte) {
	hash = st.nextHash()
	var raw []byte
	if ptype == pongPacket {
		var err error
		raw, _, err = encodePacket(st.priv, byte(ptype), data)
		if err != nil {
			panic(err)
		}
	}

	st.sendPacket(remote.ID, ingressPacket{
		remoteID:   st.sender,
		remoteAddr: st.senderAddr,
		hash:       hash,
		ev:         ptype,
		data:       data,
		rawData:    raw,
	})
	return hash
}

func (st *simTransport) sendPing(remote *Node, remoteAddr *net.UDPAddr, topics []Topic) []byte {
	hash := st.nextHash()
	st.sendPacket(remote.ID, ingressPacket{
		remoteID:   st.sender,
		remoteAddr: st.senderAddr,
		hash:       hash,
		ev:         pingPacket,
		data: &ping{
			Version:    4,
			From:       rpcEndpoint{IP: st.senderAddr.IP, UDP: uint16(st.senderAddr.Port), TCP: 30303},
			To:         rpcEndpoint{IP: remoteAddr.IP, UDP: uint16(remoteAddr.Port), TCP: 30303},
			Expiration: uint64(time.Now().Unix() + int64(expiration)),
			Topics:     topics,
		},
	})
	return hash
}

func (st *simTransport) sendFindnodeHash(remote *Node, target common.Hash) {
	st.sendPacket(remote.ID, ingressPacket{
		remoteID:   st.sender,
		remoteAddr: st.senderAddr,
		hash:       st.nextHash(),
		ev:         findnodeHashPacket,
		data: &findnodeHash{
			Target:     target,
			Expiration: uint64(time.Now().Unix() + int64(expiration)),
		},
	})
}

func (st *simTransport) sendTopicRegister(remote *Node, topics []Topic, idx int, pong []byte) {
	//fmt.Println("send", topics, pong)
	st.sendPacket(remote.ID, ingressPacket{
		remoteID:   st.sender,
		remoteAddr: st.senderAddr,
		hash:       st.nextHash(),
		ev:         topicRegisterPacket,
		data: &topicRegister{
			Topics: topics,
			Idx:    uint(idx),
			Pong:   pong,
		},
	})
}

func (st *simTransport) sendTopicNodes(remote *Node, queryHash common.Hash, nodes []*Node) {
	rnodes := make([]rpcNode, len(nodes))
	for i := range nodes {
		rnodes[i] = nodeToRPC(nodes[i])
	}
	st.sendPacket(remote.ID, ingressPacket{
		remoteID:   st.sender,
		remoteAddr: st.senderAddr,
		hash:       st.nextHash(),
		ev:         topicNodesPacket,
		data:       &topicNodes{Echo: queryHash, Nodes: rnodes},
	})
}

func (st *simTransport) sendNeighbours(remote *Node, nodes []*Node) {
	// TODO: send multiple packets
	rnodes := make([]rpcNode, len(nodes))
	for i := range nodes {
		rnodes[i] = nodeToRPC(nodes[i])
	}
	st.sendPacket(remote.ID, ingressPacket{
		remoteID:   st.sender,
		remoteAddr: st.senderAddr,
		hash:       st.nextHash(),
		ev:         neighborsPacket,
		data: &neighbors{
			Nodes:      rnodes,
			Expiration: uint64(time.Now().Unix() + int64(expiration)),
		},
	})
}

func (st *simTransport) nextHash() []byte {
	v := atomic.AddUint64(&st.hashctr, 1)
	var hash common.Hash
	binary.BigEndian.PutUint64(hash[:], v)
	return hash[:]
}

const packetLoss = 0 // 1/1000

func (st *simTransport) sendPacket(remote NodeID, p ingressPacket) {
	if rand.Int31n(1000) >= packetLoss {
		st.sim.mu.RLock()
		recipient := st.sim.nodes[remote]
		st.sim.mu.RUnlock()

		time.AfterFunc(200*time.Millisecond, func() {
			recipient.reqReadPacket(p)
		})
	}
}



