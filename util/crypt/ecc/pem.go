package ecc

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

const (
	pemBlockPrivate = "EC PRIVATE KEY"
	pemBlockPublic  = "PUBLIC KEY"
	pemCipher       = x509.PEMCipherAES256
)

// PEM marshals the PublicKey and encodes it in standard PEM format.
//
// x509.MarshalPKIXPublicKey serialises a public key to DER-encoded PKIX format.
func (pub *PublicKey) PEM() ([]byte, error) {
	marshaled, err := x509.MarshalPKIXPublicKey(pub.Key)
	if err != nil {
		return nil, err
	}
	block := &pem.Block{Type: pemBlockPublic, Bytes: marshaled}
	pemBytes := pem.EncodeToMemory(block)
	if pemBytes == nil {
		return nil, errors.New("cannot encode PEM, invalid headers")
	}
	return pemBytes, nil
}

// PEM marshals the PrivateKey and encodes it in standard PEM format. If
// password is not blank, the PEM will be encrypted using x509.PEMCipherAES256.
//
// x509.MarshalECPrivateKey marshals an EC private key into ASN.1, DER format.
//
// x509.EncryptPEMBlock returns a PEM block of the specified type holding the
// given DER-encoded data encrypted with the specified algorithm and password.
func (priv *PrivateKey) PEM(password string) ([]byte, error) {
	marshaled, err := x509.MarshalECPrivateKey(priv.Key)
	if err != nil {
		return nil, err
	}
	block := &pem.Block{Type: pemBlockPrivate, Bytes: marshaled}
	if password != "" {
		block, err = x509.EncryptPEMBlock(rand.Reader, block.Type, block.Bytes, []byte(password), pemCipher)
		if err != nil {
			return nil, err
		}
	}
	pemBytes := pem.EncodeToMemory(block)
	if pemBytes == nil {
		return nil, errors.New("cannot encode PEM, invalid headers")
	}
	return pemBytes, nil
}

// DecodePEMPublicKey decodes the PEM format for the elliptic PublicKey.
//
// pem.Decode will find the next PEM formatted block (certificate, private key
// etc) in the input. It returns that block and the remainder of the input. If
// no PEM data is found, p is nil and the whole of the input is returned in
// rest.
func DecodePEMPublicKey(pemEncodedKey []byte) (*PublicKey, error) {
	block, _ := pem.Decode(pemEncodedKey) // discard any trailing text
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing key")
	}
	pkixPublicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	ecdsaPublicKey, ok := pkixPublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("failed to cast to PublicKey")
	}
	pub := &PublicKey{Key: ecdsaPublicKey}
	return pub, nil
}

// DecodePEMPrivateKey decodes the PEM format for the elliptic PrivateKey. If
// password is not blank, the PEM will be decrypted first.
//
// pem.Decode will find the next PEM formatted block (certificate, private key
// etc) in the input. It returns that block and the remainder of the input. If
// no PEM data is found, p is nil and the whole of the input is returned in
// rest.
//
// x509.DecryptPEMBlock takes a password encrypted PEM block and the password
// used to encrypt it and returns a slice of decrypted DER encoded bytes. It
// inspects the DEK-Info header to determine the algorithm used for decryption.
// If no DEK-Info header is present, an error is returned. If an incorrect
// password is detected an IncorrectPasswordError is returned. Because of
// deficiencies in the encrypted-PEM format, it's not always possible to detect
// an incorrect password. In these cases no error will be returned but the
// decrypted DER bytes will be random noise.
func DecodePEMPrivateKey(pemEncodedKey []byte, password string) (*PrivateKey, error) {
	var err error
	block, _ := pem.Decode(pemEncodedKey) // discard any trailing text
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing key")
	}
	unencrypted := block.Bytes
	if x509.IsEncryptedPEMBlock(block) {
		if password == "" {
			return nil, errors.New("PEM is encrypted and password is blank")
		}
		unencrypted, err = x509.DecryptPEMBlock(block, []byte(password))
		if err != nil {
			return nil, err
		}
	}
	ecdsaPrivateKey, err := x509.ParseECPrivateKey(unencrypted)
	if err != nil {
		return nil, err
	}
	priv := &PrivateKey{Key: ecdsaPrivateKey}
	return priv, nil
}
