package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
)

// GenerateRSAKeyPair creates a 2048-bit RSA key pair.
func GenerateRSAKeyPair() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

// ExportPublicKey encodes a public key as a base64-encoded PKIX/SPKI DER.
func ExportPublicKey(pub *rsa.PublicKey) (string, error) {
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(der), nil
}

// ExportPrivateKey encodes a private key as a base64-encoded PKCS#8 DER.
func ExportPrivateKey(priv *rsa.PrivateKey) (string, error) {
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(der), nil
}

// ParsePublicKey decodes a base64 PKIX public key string.
func ParsePublicKey(encoded string) (*rsa.PublicKey, error) {
	der, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	pub, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return nil, err
	}
	return pub.(*rsa.PublicKey), nil
}

// ParsePrivateKey decodes a base64 PKCS#8 private key string.
func ParsePrivateKey(encoded string) (*rsa.PrivateKey, error) {
	der, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	priv, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return nil, err
	}
	return priv.(*rsa.PrivateKey), nil
}

// DecryptOAEP decrypts an RSA-OAEP encrypted ciphertext.
func DecryptOAEP(priv *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, priv, ciphertext, nil)
}
