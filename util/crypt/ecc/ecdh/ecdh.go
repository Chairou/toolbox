// Package ecdh encrypts and decrypts data using elliptic curve keys. Data
// is encrypted with AES-128-CBC with HMAC-SHA1 message tags using
// ECDHE to generate a shared key. The P256 curve is chosen in
// keeping with the use of AES-128 for encryption.
package ecdh

import (
	"bytes"
	"crypto/aes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"errors"
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
		return nil, err
	}
	x, y := pub.Curve.ScalarMult(pub.X, pub.Y, ephemeral.D.Bytes())
	x, y = pub.Double(x, y)
	if x == nil {
		return nil, errors.New("failed to generate encryption key")
	}
	//fmt.Println("X0 :", x)
	byteSum := JoinBytes(x.Bytes(), y.Bytes())
	//shared := sha512.Sum512(x.Bytes())
	shared := sha512.Sum512(byteSum)
	//fmt.Println("shared:", shared)
	iv, err := symcrypt.MakeRandom(16)
	if err != nil {
		return nil, err
	}

	paddedIn := padding.AddPaddingB(in)
	ct, err := symcrypt.EncryptCBC(paddedIn, iv, getKeyFromX(x.Bytes(), y.Bytes(), iv, shared))
	if err != nil {
		return nil, err
	}

	ephPub := elliptic.MarshalCompressed(pub.Curve, ephemeral.PublicKey.X, ephemeral.PublicKey.Y)
	//out = make([]byte, 1+len(ephPub)+16)

	num, _ := symcrypt.RandomInRange(13, 256)
	randBytes, _ := symcrypt.MakeRandom(int(num))

	out = make([]byte, num+1+1+len(ephPub)+16)
	out[0] = byte(num) - 5
	copy(out[1:num], randBytes)
	out[0+num+1] = byte(len(ephPub))
	copy(out[1+num+1:], ephPub)
	copy(out[1+num+1+len(ephPub):], iv)
	out = append(out, ct...)
	//fmt.Println("RAW BYTE:", out)

	//out = make([]byte, 1+len(ephPub)+16)
	//out[0] = byte(len(ephPub))
	//copy(out[1:], ephPub)
	//copy(out[1+len(ephPub):], iv)
	//out = append(out, ct...)

	h := hmac.New(sha512.New, shared[:])
	h.Write(iv)
	h.Write(ct)
	out = h.Sum(out)
	return out, nil
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
	//fmt.Println("x1:", x)
	x, y = priv.Double(x, y)
	//fmt.Println("x2:", x)

	if x == nil {
		return nil, errors.New("failed to generate encryption key")
	}
	byteSum := JoinBytes(x.Bytes(), y.Bytes())

	//shared := sha512.Sum512(x.Bytes())
	shared := sha512.Sum512(byteSum)

	tagStart := len(ct) - sha512.Size
	h := hmac.New(sha512.New, shared[:])
	h.Write(ct[:tagStart])
	mac := h.Sum(nil)
	if !hmac.Equal(mac, ct[tagStart:]) {
		return nil, errors.New("invalid MAC")
	}
	iv := ct[:aes.BlockSize]
	paddedOut, err := symcrypt.DecryptCBC(ct[aes.BlockSize:tagStart], iv, getKeyFromX(x.Bytes(), y.Bytes(), iv, shared))
	if err != nil {
		return
	}
	out, err = padding.RemovePadding(paddedOut)
	return
}

func JoinBytes(slices ...[]byte) []byte {
	return bytes.Join(slices, []byte{})
}

func getKeyFromShared(iv []byte, shared [64]byte) []byte {
	startNum := iv[0] % 32
	return shared[startNum : startNum+32]
}

func getKeyFromX(x []byte, y []byte, iv []byte, shared [64]byte) []byte {
	xPosition := shared[4] % 64
	ivPosition := shared[7] % 16
	yPosition := shared[47] % 64
	startNum := (x[xPosition] + iv[ivPosition] + y[yPosition]) % 32
	return shared[startNum : startNum+32]
}
