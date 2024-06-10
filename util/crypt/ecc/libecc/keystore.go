package main

import (
	"C"
	"github.com/Chairou/toolbox/util/crypt/ecc"
	"sync"
)

// keyStore has Go pointers to keys, and allows C clients
// to use the keys by ID. The slices are never shrunk.
// Never use index 0; this lib returns zero values to indicate errors.
type keyStore struct {
	pub       []*ecc.PublicKey
	pubMutex  sync.Mutex
	priv      []*ecc.PrivateKey
	privMutex sync.Mutex
}

func (ks *keyStore) addPubic(key *ecc.PublicKey) (pubID C.int) {
	ks.pubMutex.Lock()
	defer ks.pubMutex.Unlock()
	ks.pub = append(ks.pub, key)
	pubID = C.int(len(ks.pub) - 1)
	return
}

func (ks *keyStore) addPrivate(key *ecc.PrivateKey) (privID C.int) {
	ks.privMutex.Lock()
	defer ks.privMutex.Unlock()
	ks.priv = append(ks.priv, key)
	privID = C.int(len(ks.priv) - 1)
	return
}

func (ks *keyStore) getPubic(pubID C.int) (key *ecc.PublicKey) {
	ks.pubMutex.Lock()
	defer ks.pubMutex.Unlock()
	id := int(pubID)
	if id < 1 || id >= len(ks.pub) {
		return
	}
	key = ks.pub[id]
	return
}

func (ks *keyStore) getPrivate(privID C.int) (key *ecc.PrivateKey) {
	ks.privMutex.Lock()
	defer ks.privMutex.Unlock()
	id := int(privID)
	if id < 1 || id >= len(ks.priv) {
		return
	}
	key = ks.priv[id]
	return
}

func (ks *keyStore) delPubic(pubID C.int) {
	ks.pubMutex.Lock()
	defer ks.pubMutex.Unlock()
	id := int(pubID)
	if id < 1 || id >= len(ks.pub) {
		return
	}
	ks.pub[id] = nil
	return
}

func (ks *keyStore) delPrivate(privID C.int) {
	ks.privMutex.Lock()
	defer ks.privMutex.Unlock()
	id := int(privID)
	if id < 1 || id >= len(ks.priv) {
		return
	}
	ks.priv[id] = nil
	return
}
