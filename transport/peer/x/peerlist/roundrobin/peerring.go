// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package roundrobin

import (
	"container/ring"
	"sync"

	"go.uber.org/yarpc/transport"
	"go.uber.org/yarpc/transport/internal/errors"
)

// NewPeerRing creates a new PeerRing with an initial capacity
func NewPeerRing(capacity int) *PeerRing {
	return &PeerRing{
		peerToNode: make(map[string]*ring.Ring, capacity),
	}
}

// PeerRing provides a safe way to interact (Add/Remove/Get) with a potentially
// changing list of peer objects
type PeerRing struct {
	lock sync.Mutex

	peerToNode map[string]*ring.Ring
	nextNode   *ring.Ring
}

// Add a transport.Peer to the end of the PeerRing, if the ring is empty
// it initializes the nextNode marker
func (pr *PeerRing) Add(peer transport.Peer) error {
	pr.lock.Lock()
	defer pr.lock.Unlock()

	if _, ok := pr.peerToNode[peer.Identifier()]; ok {
		// Peer Already in ring, ignore the add
		return errors.ErrPeerAddAlreadyInList(peer.Identifier())
	}

	newNode := newPeerRingNode(peer)
	pr.peerToNode[peer.Identifier()] = newNode

	if pr.nextNode == nil {
		// Empty ring, add the first node
		pr.nextNode = newNode
	} else {
		// Push the node to the ring
		pushBeforeNode(pr.nextNode, newNode)
	}
	return nil
}

func newPeerRingNode(peer transport.Peer) *ring.Ring {
	newNode := ring.New(1)
	newNode.Value = peer
	return newNode
}

// Remove a peer PeerIdentifier from the PeerRing, if the PeerID is not
// in the ring return an error
func (pr *PeerRing) Remove(pid transport.PeerIdentifier) error {
	pr.lock.Lock()
	defer pr.lock.Unlock()

	node, ok := pr.peerToNode[pid.Identifier()]
	if !ok {
		// Peer doesn't exist in the list
		return errors.ErrPeerRemoveNotInList(pid.Identifier())
	}

	pr.popNode(node)

	return nil
}

// RemoveAll pops all the peers from the ring and returns them in a list
func (pr *PeerRing) RemoveAll() []transport.Peer {
	pr.lock.Lock()
	defer pr.lock.Unlock()

	peers := make([]transport.Peer, 0, len(pr.peerToNode))
	for _, node := range pr.peerToNode {
		peers = append(peers, pr.popNode(node))
	}
	return peers
}

// Must be run inside a mutex.Lock()
func (pr *PeerRing) popNode(node *ring.Ring) transport.Peer {
	p := getPeerForRingNode(node)

	if isLastRingNode(node) {
		pr.nextNode = nil
	} else {
		if pr.isNextNode(node) {
			pr.nextNode = pr.nextNode.Next()
		}
		popNodeFromRing(node)
	}

	// Remove the node from our node map
	delete(pr.peerToNode, p.Identifier())

	return p
}

// Must be run inside a mutex.Lock()
func (pr *PeerRing) isNextNode(node *ring.Ring) bool {
	return pr.nextNode == node
}

// Next returns the next peer in the ring, or nil if there is no peer in the ring
// after it has the next peer, it increments the nextPeer marker in the ring
func (pr *PeerRing) Next() transport.Peer {
	pr.lock.Lock()
	defer pr.lock.Unlock()

	if pr.nextNode == nil {
		return nil
	}

	p := getPeerForRingNode(pr.nextNode)

	pr.nextNode = pr.nextNode.Next()

	return p
}

func getPeerForRingNode(rNode *ring.Ring) transport.Peer {
	return rNode.Value.(transport.Peer)
}

func isLastRingNode(rNode *ring.Ring) bool {
	return rNode.Next() == rNode
}

func popNodeFromRing(rNode *ring.Ring) {
	rNode.Prev().Unlink(1)
}

func pushBeforeNode(curNode, newNode *ring.Ring) {
	curNode.Prev().Link(newNode)
}