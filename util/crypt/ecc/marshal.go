package ecc

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"errors"
)

// Marshal converts the PublicKey to a byte slice.
//
// elliptic.Marshal converts a point into the uncompressed form specified in
// section 4.3.6 of ANSI X9.62.
func (pub *PublicKey) Marshal() []byte {
	if pub.Key == nil {
		return []byte{}
	}
	return elliptic.Marshal(pub.Key.Curve, pub.Key.X, pub.Key.Y)
}

// Marshal converts a PrivateKey to a byte slice.
//
// MarshalECPrivateKey marshals an EC private key into ASN.1, DER format.
func (priv *PrivateKey) Marshal() ([]byte, error) {
	if priv.Key == nil {
		return []byte{}, errors.New("malformed PrivateKey, Key == nil")
	}
	return x509.MarshalECPrivateKey(priv.Key)
}

// UnmarshalPublicKey loads a PublicKey from a marshalled byte slice.
//
// elliptic.Unmarshal converts a point, serialized by Marshal, into an x, y
// pair. It is an error if the point is not in uncompressed form or is not on
// the curve. On error, x = nil.
func UnmarshalPublicKey(curve elliptic.Curve, marshalledKey []byte) *PublicKey {
	if curve == nil || len(marshalledKey) == 0 {
		return nil
	}
	x, y := elliptic.Unmarshal(curve, marshalledKey)
	ecdsaPublicKey := &ecdsa.PublicKey{Curve: curve, X: x, Y: y}
	pub := &PublicKey{Key: ecdsaPublicKey}
	return pub
}

// UnmarshalPrivateKey loads a PrivateKey from a marshalled byte slice.
//
// x509.ParseECPrivateKey parses an ASN.1 Elliptic Curve Private Key Structure.
func UnmarshalPrivateKey(marshalledKey []byte) (*PrivateKey, error) {
	ecdsaPrivateKey, err := x509.ParseECPrivateKey(marshalledKey)
	if err != nil {
		return nil, err
	}
	priv := &PrivateKey{Key: ecdsaPrivateKey}
	return priv, nil
}
