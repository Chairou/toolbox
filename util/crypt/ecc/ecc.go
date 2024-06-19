// Package ecc makes public key elliptic curve cryptography easier to use.
// Cryptographic functions are from "crypto/ecdsa" and
// "github.com/cloudflare/redoctober/ecdh".
//
//	Sign/Verify                // authentication by signing a hash
//	SignMessage/VerifyMessage  // authentication by hashing a message and signing the hash SHA256
//	Encrypt/Decrypt            // encrypts with ephemeral symmetrical key AES-128-CBC-HMAC-SHA1
//	Seal/Open                  // both Sign and Encrypt, then Decrypt and Verify
//
//	Marshal/Unmarshal          // convert keys to and from []byte slices
//	PEM/DecodePEM              // keys in PEM file format, can be encrypted with a password
//
// Packages ecdh, padding and symcrypt are copied from redoctober into this
// package so it works with go get. Only a few minor changes were made: package
// import paths, an error check and a comment added. Run ./diff_redoctober.sh to
// see all changes.
package ecc

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"errors"
	"fmt"
	"github.com/Chairou/toolbox/util/crypt/ecc/ecdh"
	"math/big"
	"sync"
)

// signatures are ASN.1 encoded (r, s *big.Int), minimum is 1 byte each for:
// sequence, length, integer, length, value, integer, length, value
const signatureMinLen int = 8

// sigRS is used with asn1.Marshal() and asn1.Unmarshal()
type sigRS struct {
	R *big.Int
	S *big.Int
}

var (
	messageHash    = sha256.Sum256 // the hash used by Seal/Open
	mutexEcdhCurve = sync.Mutex{}  // set ecdh.Curve for Encrypt/Decrypt
)

// GenerateKeys creates a new public and private key pair. Note that the public
// key is double the size of the private key. Using elliptic.P256() would make
// the private key 32 bytes and the public key 64 bytes.
//
//	pub, priv, err := ecc.GenerateKeys(elliptic.P256())
func GenerateKeys(curve elliptic.Curve) (pub *PublicKey, priv *PrivateKey, err error) {
	defer func() {
		if recover() != nil {
			err = errors.New("panic in ecdsa.GenerateKey; invalid curve")
		}
	}()
	ecdsaPrivateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return
	}
	pub = &PublicKey{Key: &ecdsaPrivateKey.PublicKey}
	priv = &PrivateKey{Key: ecdsaPrivateKey}
	return
}

// PublicKey is the elliptic public key.
type PublicKey struct {
	Key *ecdsa.PublicKey
}

// Verify checks the signature was created by the PrivateKey that goes with this
// PublicKey. Returns true if valid.
//
// Signatures are created by the sender calling PrivateKey.Sign(hash) where the
// hash is of a larger message. The recipient uses the senders public key to
// check the senders signature. The recipient must hash the larger message (with
// the same algorithm) and verify the hash with the signature.
func (pub *PublicKey) Verify(hash, signature []byte) (bool, error) {
	if len(hash) == 0 || len(signature) < signatureMinLen {
		return false, errors.New("length of hash or signature []byte slice is too short")
	}
	rs := sigRS{}
	_, err := asn1.Unmarshal(signature, &rs) // ignore "rest" of the bytes
	if rs.R == nil || rs.S == nil || err != nil {
		return false, err
	}
	return ecdsa.Verify(pub.Key, hash, rs.R, rs.S), nil
}

// VerifyMessage will hash the message with Sha256, and then verify the
// signature.
func (pub *PublicKey) VerifyMessage(message, signature []byte) (verified bool, err error) {
	hash := messageHash(message)
	verified, err = pub.Verify(hash[:], signature)
	return
}

// Encrypt secures and authenticates its input with the public key, using ECDHE
// with AES-128-CBC-HMAC-SHA1.
func (pub *PublicKey) Encrypt(message []byte) (encrypted []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			msg := fmt.Sprint("panic in ecdh.Encrypt:", e)
			err = errors.New(msg)
		}
	}()
	// lock ecdh.Curve because the ephemeral key needs the same Curve as pub
	mutexEcdhCurve.Lock()
	defer mutexEcdhCurve.Unlock()
	ecdh.Curve = func() elliptic.Curve {
		//t := reflect.TypeOf(pub.Key.Curve)
		//fmt.Println("Curve type: ", t)
		return pub.Key.Curve
	}

	encrypted, err = ecdh.Encrypt(pub.Key, message)
	return
}

// PrivateKey is the elliptic private key.
type PrivateKey struct {
	Key *ecdsa.PrivateKey
}

// Sign uses the PrivateKey to create a signature of a hash (which should be the
// result of hashing a larger message.) Use Verify() after Sign.
//
// If the hash is longer than the bit-length of the private key's curve order,
// the hash will be truncated to that length.
func (priv *PrivateKey) Sign(hash []byte) (signature []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			msg := fmt.Sprint("panic in ecdsa.Sign:", e)
			err = errors.New(msg)
		}
	}()
	r, s, err := ecdsa.Sign(rand.Reader, priv.Key, hash)
	if r == nil || s == nil || err != nil {
		return
	}
	signature, err = asn1.Marshal(sigRS{R: r, S: s})
	return
}

// SignMessage will hash the message with Sha256 and sign the hash. Returns a
// signature and error. Use VerifyMessage() after SignMessage.
func (priv *PrivateKey) SignMessage(message []byte) (signature []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			msg := fmt.Sprint("panic in messageHash:", e)
			err = errors.New(msg)
		}
	}()
	hash := messageHash(message)
	signature, err = priv.Sign(hash[:])
	return
}

// Decrypt authenticates and recovers the original message from its input using
// the private key and the ephemeral key included in the message.
func (priv *PrivateKey) Decrypt(encrypted []byte) (message []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			msg := fmt.Sprint("panic in ecdh.Decrypt:", e)
			err = errors.New(msg)
		}
	}()
	// lock ecdh.Curve because the ephemeral key needs the same Curve as pub
	mutexEcdhCurve.Lock()
	defer mutexEcdhCurve.Unlock()
	ecdh.Curve = func() elliptic.Curve { return priv.Key.Curve }
	message, err = ecdh.Decrypt(priv.Key, encrypted)
	return
}

// Seal authenticates and encrypts by calling SignMessage() and Encrypt().
// Use Open() after Seal.
func (priv *PrivateKey) Seal(message []byte, to *PublicKey) (sealed []byte, err error) {
	signature, err := priv.SignMessage(message)
	if len(signature) == 0 || err != nil {
		return nil, err
	}
	// format is signature + message
	unencrypted := make([]byte, 0)
	unencrypted = append(unencrypted, signature...)
	unencrypted = append(unencrypted, message...)
	sealed, err = to.Encrypt(unencrypted)
	return
}

// Open decrypts and authenticates by calling Decrypt() and VerifyMessage().
// Open is used after Seal().
func (priv *PrivateKey) Open(sealed []byte, from *PublicKey) ([]byte, error) {
	decrypted, err := priv.Decrypt(sealed)
	if err != nil {
		return nil, err
	}
	if len(decrypted) < signatureMinLen {
		return nil, errors.New("signature is too short")
	}
	// format is signature + message
	// message is bytes after the signature; what Unmarshal returns as "rest"
	message, err := asn1.Unmarshal(decrypted, &sigRS{})
	if err != nil {
		return nil, err
	}
	// signature is from zero to the message
	signature := decrypted[:len(decrypted)-len(message)]
	verified, err := from.VerifyMessage(message, signature)
	if err != nil {
		return nil, err
	}
	if !verified {
		return nil, errors.New("signature could not be verified")
	}
	return message, nil
}
