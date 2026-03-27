// Package ssh provides SSH key generation utilities
package ssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"

	"golang.org/x/crypto/ssh"
)

// GenerateKeyPair generates an Ed25519 SSH keypair
// Returns private key in OpenSSH PEM format and public key in authorized_keys format
func GenerateKeyPair() (privateKey string, publicKey string, err error) {
	// Generate Ed25519 keypair
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}

	// Encode private key (OpenSSH PEM format)
	privateKeyPEM, err := ssh.MarshalPrivateKey(priv, "")
	if err != nil {
		return "", "", err
	}

	privateBytes := pem.EncodeToMemory(privateKeyPEM)

	// Encode public key (authorized_keys format)
	pubKey, err := ssh.NewPublicKey(pub)
	if err != nil {
		return "", "", err
	}

	pubBytes := ssh.MarshalAuthorizedKey(pubKey)

	return string(privateBytes), string(pubBytes), nil
}
