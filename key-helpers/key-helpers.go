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

// Takes a public key that was converted into a string and turns it back into a public key
// Returns the key
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

// Takes a private key that was converted into a string and turns it back into a private key
// Returns the key
func PrivateKeyStringToKey(key string) *ecdsa.PrivateKey {
	privateKeyBytesRestored, _ := pem.Decode([]byte(key))
	x509Encoded := privateKeyBytesRestored.Bytes
	privateKey, err := x509.ParseECPrivateKey(x509Encoded)
	if err != nil {
		log.Fatal(err)
	}
	return privateKey
}

// Checks that the public and private keys that were restored from strings are still valid
// Returns true if the key restore was successful
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

// Encodes a public/private keypair to strings for easier storage and sending
// https://stackoverflow.com/questions/21322182/how-to-store-ecdsa-private-key-in-go
func Encode(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey) (privateKeyString string, publicKeyString string) {
	x509Encoded, _ := x509.MarshalECPrivateKey(privateKey)
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})

	x509EncodedPub, _ := x509.MarshalPKIXPublicKey(publicKey)
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})

	return string(pemEncoded), string(pemEncodedPub)
}

// Generates a new public/private keypair and returns them
func GenerateKeys() (*ecdsa.PublicKey, *ecdsa.PrivateKey){
	testPrivateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	testPublicKey := &testPrivateKey.PublicKey
	if err != nil {
		log.Fatal(err)
	}
	return testPublicKey, testPrivateKey
}

// Converts a public key to string
// Returns the string-encoded public key
func PubKeyToString(key ecdsa.PublicKey) string {
	return string(elliptic.Marshal(key.Curve, key.X, key.Y))
}

// Decodes keys the way they are encoded by the above function
func StringToPubKey(keyString string) ecdsa.PublicKey {
	x, y := elliptic.Unmarshal(elliptic.P384(), []byte(keyString))
	key := ecdsa.PublicKey{elliptic.P384(), x, y}
	return key
}
