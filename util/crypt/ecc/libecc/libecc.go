// libecc is a C wrapper for package ecc.
// Nil pointers and zero values are used to indicate errors,
// and error messages are not returned.
// Go byte slices are passed as *C.ecc_bytes
package main

import (
	"crypto/elliptic"
	"github.com/Chairou/toolbox/util/crypt/ecc"
	"unsafe"
)

/*
#include <stdlib.h>

typedef struct { int pubID; int privID; } ecc_keypair;

static ecc_keypair* ecc_keypair_new(int pubID, int privID) {
	ecc_keypair* kp = malloc(sizeof(ecc_keypair));
	kp->pubID = pubID;
	kp->privID = privID;
	return kp;
}

typedef struct { int n; unsigned char* data; } ecc_bytes;

static ecc_bytes* ecc_bytes_new(int n, void* data) {
	ecc_bytes* r = malloc(sizeof(ecc_bytes));
	r->n = n;
	r->data = (unsigned char*)data;
	return r;
}
*/
import "C"

func eccBytes(b []byte) *C.ecc_bytes {
	n := C.int(len(b))
	data := C.CBytes(b)
	return C.ecc_bytes_new(n, data)
}

func eccBytes2Slice(b *C.ecc_bytes) []byte {
	return C.GoBytes(unsafe.Pointer(b.data), b.n)
}

// keys is the single global keyStore
// Never use index 0; this lib returns zero values to indicate errors.
var keys = keyStore{
	pub:  make([]*ecc.PublicKey, 1),
	priv: make([]*ecc.PrivateKey, 1),
}

// Go requires libraries to have an empty main() function.
func main() {}

func size2curve(size int) *elliptic.Curve {
	var c elliptic.Curve
	switch size {
	case 224:
		c = elliptic.P224()
	case 256:
		c = elliptic.P256()
	case 384:
		c = elliptic.P384()
	case 521:
		c = elliptic.P521()
	}
	return &c
}

///////////////////////////// ecc functions ////////////////////////////////////

// ecc_generate_keys makes new elliptic keys.
// p0 = size [224,256,384,521]
// error will return 0
//
//export ecc_generate_keys
func ecc_generate_keys(size C.int) (keypair *C.ecc_keypair) {
	c := size2curve(int(size))
	if *c == nil {
		return //invalid key size
	}
	pub, priv, err := ecc.GenerateKeys(*c)
	if err != nil {
		return //error creating keys
	}
	pubID := keys.addPubic(pub)
	privID := keys.addPrivate(priv)
	keypair = C.ecc_keypair_new(pubID, privID)
	return
}

// p0 = pubID
// p1 = hash
// p2 = signature
// r0 = verified (zero is false, 1 is true)
//
//export ecc_pub_verify
func ecc_pub_verify(pubID C.int, hash, signature *C.ecc_bytes) (verified C.int) {
	pub := keys.getPubic(pubID)
	if pub == nil {
		return
	}
	h := eccBytes2Slice(hash)
	s := eccBytes2Slice(signature)
	v, err := pub.Verify(h, s)
	if v == false || err != nil {
		return
	}
	verified = 1
	return
}

//export ecc_pub_verifymessage
func ecc_pub_verifymessage(pubID C.int, message, signature *C.ecc_bytes) (verified C.int) {
	pub := keys.getPubic(pubID)
	if pub == nil {
		return
	}
	m := eccBytes2Slice(message)
	s := eccBytes2Slice(signature)
	v, err := pub.VerifyMessage(m, s)
	if v == false || err != nil {
		return
	}
	verified = 1
	return
}

//export ecc_pub_encrypt
func ecc_pub_encrypt(pubID C.int, message *C.ecc_bytes) (encrypted *C.ecc_bytes) {
	pub := keys.getPubic(pubID)
	if pub == nil {
		return
	}
	m := eccBytes2Slice(message)
	e, err := pub.Encrypt(m)
	if len(e) == 0 || err != nil {
		return
	}
	encrypted = eccBytes(e)
	return
}

//export ecc_priv_sign
func ecc_priv_sign(privID C.int, hash *C.ecc_bytes) (signature *C.ecc_bytes) {
	priv := keys.getPrivate(privID)
	if priv == nil {
		return
	}
	h := eccBytes2Slice(hash)
	s, err := priv.Sign(h)
	if len(s) == 0 || err != nil {
		return
	}
	signature = eccBytes(s)
	return
}

//export ecc_priv_signmessage
func ecc_priv_signmessage(privID C.int, message *C.ecc_bytes) (signature *C.ecc_bytes) {
	priv := keys.getPrivate(privID)
	if priv == nil {
		return
	}
	m := eccBytes2Slice(message)
	s, err := priv.SignMessage(m)
	if len(s) == 0 || err != nil {
		return
	}
	signature = eccBytes(s)
	return
}

//export ecc_priv_decrypt
func ecc_priv_decrypt(privID C.int, encrypted *C.ecc_bytes) (message *C.ecc_bytes) {
	priv := keys.getPrivate(privID)
	if priv == nil {
		return
	}
	e := eccBytes2Slice(encrypted)
	m, err := priv.Decrypt(e)
	if len(m) == 0 || err != nil {
		return
	}
	message = eccBytes(m)
	return
}

//export ecc_priv_seal
func ecc_priv_seal(privID, toPubID C.int, message *C.ecc_bytes) (sealed *C.ecc_bytes) {
	priv := keys.getPrivate(privID)
	if priv == nil {
		return
	}
	pub := keys.getPubic(toPubID)
	if pub == nil {
		return
	}
	m := eccBytes2Slice(message)
	s, err := priv.Seal(m, pub)
	if len(s) == 0 || err != nil {
		return
	}
	sealed = eccBytes(s)
	return
}

//export ecc_priv_open
func ecc_priv_open(privID, fromPubID C.int, sealed *C.ecc_bytes) (opened *C.ecc_bytes) {
	priv := keys.getPrivate(privID)
	if priv == nil {
		return
	}
	pub := keys.getPubic(fromPubID)
	if pub == nil {
		return
	}
	s := eccBytes2Slice(sealed)
	o, err := priv.Open(s, pub)
	if len(o) == 0 || err != nil {
		return
	}
	opened = eccBytes(o)
	return
}

///////////////////////////// marshal //////////////////////////////////////////

// ecc_pub_marshal ...
//
//export ecc_pub_marshal
func ecc_pub_marshal(pubID C.int) (marshalled *C.ecc_bytes) {
	pub := keys.getPubic(pubID)
	if pub == nil {
		return
	}
	marshalled = eccBytes(pub.Marshal())
	return
}

// ecc_priv_marshal ...
//
//export ecc_priv_marshal
func ecc_priv_marshal(privID C.int) (marshalled *C.ecc_bytes) {
	priv := keys.getPrivate(privID)
	if priv == nil {
		return
	}
	m, err := priv.Marshal()
	if err != nil {
		return
	}
	marshalled = eccBytes(m)
	return
}

//export ecc_pub_unmarshal
func ecc_pub_unmarshal(size C.int, marshalled *C.ecc_bytes) (pubID C.int) {
	c := size2curve(int(size))
	if *c == nil {
		return //invalid key size
	}
	m := eccBytes2Slice(marshalled)
	pub := ecc.UnmarshalPublicKey(*c, m)
	if pub == nil {
		return
	}
	pubID = keys.addPubic(pub)
	return
}

//export ecc_priv_unmarshal
func ecc_priv_unmarshal(marshalled *C.ecc_bytes) (keypair *C.ecc_keypair) {
	m := eccBytes2Slice(marshalled)
	priv, err := ecc.UnmarshalPrivateKey(m)
	if priv == nil || err != nil {
		return
	}
	pub := &ecc.PublicKey{Key: &priv.Key.PublicKey}
	pubID := keys.addPubic(pub)
	privID := keys.addPrivate(priv)
	keypair = C.ecc_keypair_new(pubID, privID)
	return
}

///////////////////////////// pem //////////////////////////////////////////////

//export ecc_pub_pem
func ecc_pub_pem(pubID C.int) (pem *C.char) {
	pub := keys.getPubic(pubID)
	if pub == nil {
		return
	}
	p, err := pub.PEM()
	if len(p) == 0 || err != nil {
		return
	}
	pem = C.CString(string(p))
	return
}

//export ecc_priv_pem
func ecc_priv_pem(privID C.int, password *C.char) (pem *C.char) {
	priv := keys.getPrivate(privID)
	if priv == nil {
		return
	}
	pass := C.GoString(password)
	p, err := priv.PEM(pass)
	if len(p) == 0 || err != nil {
		return
	}
	pem = C.CString(string(p))
	return
}

//export ecc_pub_decode_pem
func ecc_pub_decode_pem(pem *C.char) (pubID C.int) {
	pub, err := ecc.DecodePEMPublicKey([]byte(C.GoString(pem)))
	if pub == nil || err != nil {
		return
	}
	pubID = keys.addPubic(pub)
	return
}

//export ecc_priv_decode_pem
func ecc_priv_decode_pem(pem *C.char, password *C.char) (keypair *C.ecc_keypair) {
	pass := C.GoString(password)
	priv, err := ecc.DecodePEMPrivateKey([]byte(C.GoString(pem)), pass)
	if priv == nil || err != nil {
		return
	}
	pub := &ecc.PublicKey{Key: &priv.Key.PublicKey}
	pubID := keys.addPubic(pub)
	privID := keys.addPrivate(priv)
	keypair = C.ecc_keypair_new(pubID, privID)
	return
}
