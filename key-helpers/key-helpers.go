package key_helpers

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/pem"
	"crypto/x509"
	"log"
	"crypto/elliptic"
	"math/big"
)

type ToVerify struct{
	R *big.Int
	S *big.Int
}
func PublicKeyStringToKey(key string) *ecdsa.PublicKey {
	publicKeyBytesRestored, _ := pem.Decode([]byte(key))
	x509Encoded := publicKeyBytesRestored.Bytes
	publicKey, err := x509.ParsePKIXPublicKey(x509Encoded)
	if err != nil {
		log.Fatal("HERE", err)
	}
	p, ok := publicKey.(*ecdsa.PublicKey)
	if ok {
		return p
	} else {
		log.Fatal("Not of type ecdsa.PublicKey")
		return nil
	}
}

func PrivateKeyStringToKey(key string) *ecdsa.PrivateKey {
	privateKeyBytesRestored, _ := pem.Decode([]byte(key))
	x509Encoded := privateKeyBytesRestored.Bytes
	privateKey, err := x509.ParseECPrivateKey(x509Encoded)
	if err != nil {
		log.Fatal(err)
	}
	return privateKey
}

func CheckKeyRestore(publicKey *ecdsa.PublicKey, privateKey *ecdsa.PrivateKey) bool {
	data := []byte("data")
	// Signing by private key
	r, s, _ := ecdsa.Sign(rand.Reader, privateKey, data)

	// Verifying against public key
	if !ecdsa.Verify(publicKey, data, r, s) {
		return false
	}
	return true
}

// https://stackoverflow.com/questions/21322182/how-to-store-ecdsa-private-key-in-go
func Encode(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey) (privateKeyString string, publicKeyString string) {
	x509Encoded, _ := x509.MarshalECPrivateKey(privateKey)
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})

	x509EncodedPub, _ := x509.MarshalPKIXPublicKey(publicKey)
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})

	return string(pemEncoded), string(pemEncodedPub)
}

func EncodePubKey(publicKey *ecdsa.PublicKey) (publicKeyString string) {
	x509EncodedPub, _ := x509.MarshalPKIXPublicKey(publicKey)
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})

	return string(pemEncodedPub)
}

func GenerateKeys() (*ecdsa.PublicKey, *ecdsa.PrivateKey){
	testPrivateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	testPublicKey := &testPrivateKey.PublicKey
	if err != nil {
		log.Fatal(err)
	}
	return testPublicKey, testPrivateKey
}