// Package ecdh encrypts and decrypts data using elliptic curve keys. Data
// is encrypted with AES-128-CBC with HMAC-SHA1 message tags using
// ECDHE to generate a shared key. The P256 curve is chosen in
// keeping with the use of AES-128 for encryption.
package ecdh

import (
	"crypto/aes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"fmt"
	"github.com/Chairou/toolbox/util/crypt/ecc/padding"
	"github.com/Chairou/toolbox/util/crypt/ecc/symcrypt"
)

// Curve is a default elliptic.P256
var Curve = elliptic.P521

// Encrypt secures and authenticates its input using the public key
// using ECDHE with AES-128-CBC-HMAC-SHA1.
func Encrypt(pub *ecdsa.PublicKey, in []byte) (out []byte, err error) {
	ephemeral, err := ecdsa.GenerateKey(Curve(), rand.Reader)
	if err != nil {
		return
	}
	x, y := pub.Curve.ScalarMult(pub.X, pub.Y, ephemeral.D.Bytes())
	x, _ = pub.Double(x, y)
	if x == nil {
		return nil, errors.New("Failed to generate encryption key")
	}
	shared := sha512.Sum512(x.Bytes())
	fmt.Println("shared:", shared)
	iv, err := symcrypt.MakeRandom(16)
	if err != nil {
		return
	}

	paddedIn := padding.AddPaddingB(in)
	ct, err := symcrypt.EncryptCBC(paddedIn, iv, shared[32:64])
	if err != nil {
		return
	}

	ephPub := elliptic.MarshalCompressed(pub.Curve, ephemeral.PublicKey.X, ephemeral.PublicKey.Y)
	//out = make([]byte, 1+len(ephPub)+16)

	num, _ := symcrypt.RandomInRange(10, 200)
	randBytes, _ := symcrypt.MakeRandom(int(num))

	out = make([]byte, num+1+1+len(ephPub)+16)
	out[0] = byte(num) - 5
	copy(out[1:num], randBytes)
	out[0+num+1] = byte(len(ephPub))
	copy(out[1+num+1:], ephPub)
	copy(out[1+num+1+len(ephPub):], iv)
	out = append(out, ct...)
	fmt.Println("RAW BYTE:", out)

	//out = make([]byte, 1+len(ephPub)+16)
	//out[0] = byte(len(ephPub))
	//copy(out[1:], ephPub)
	//copy(out[1+len(ephPub):], iv)
	//out = append(out, ct...)

	h := hmac.New(sha512.New, shared[:])
	h.Write(iv)
	h.Write(ct)
	out = h.Sum(out)
	return
}

// Decrypt authenticates and recovers the original message from
// its input using the private key and the ephemeral key included in
// the message.
func Decrypt(priv *ecdsa.PrivateKey, in []byte) (out []byte, err error) {
	ranLen := int(in[0]) + 5
	ephLen := int(in[0+1+ranLen])
	ephPub := in[1+ranLen+1 : 1+ranLen+1+ephLen]
	ct := in[1+ranLen+1+ephLen:]
	if len(ct) < (sha512.Size + aes.BlockSize) {
		return nil, errors.New("Invalid ciphertext")
	}

	x, y := elliptic.UnmarshalCompressed(Curve(), ephPub)

	// CHANGE from redoctober
	// panic: runtime error: invalid memory address or nil pointer dereference
	if x == nil || y == nil {
		return nil, errors.New("ecdh: failed to unmarshal ephemeral key")
	}
	// END CHANGE

	//ok := Curve().IsOnCurve(x, y) // Rejects the identity point too.
	//if x == nil || !ok {
	//	return nil, errors.New("Invalid public key")
	//}

	x, y = priv.Curve.ScalarMult(x, y, priv.D.Bytes())
	fmt.Println("x1:", x)
	x, _ = priv.Double(x, y)
	fmt.Println("x2:", x)

	if x == nil {
		return nil, errors.New("Failed to generate encryption key")
	}
	shared := sha512.Sum512(x.Bytes())

	tagStart := len(ct) - sha512.Size
	h := hmac.New(sha512.New, shared[:])
	h.Write(ct[:tagStart])
	mac := h.Sum(nil)
	if !hmac.Equal(mac, ct[tagStart:]) {
		return nil, errors.New("Invalid MAC")
	}

	paddedOut, err := symcrypt.DecryptCBC(ct[aes.BlockSize:tagStart], ct[:aes.BlockSize], shared[32:64])
	if err != nil {
		return
	}
	out, err = padding.RemovePadding(paddedOut)
	return
}
