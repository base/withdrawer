package signer

import (
	"crypto/ecdsa"
	"errors"
	"math/big"

	opcrypto "github.com/ethereum-optimism/optimism/op-service/crypto"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// ecdsaSigner represents an ECDSA signer.
type ecdsaSigner struct {
	// privKey is encapsulated to prevent external mutation or accidental exposure.
	privKey *ecdsa.PrivateKey
}

// NewECDSASigner validates the private key and returns a new signer instance.
func NewECDSASigner(pk *ecdsa.PrivateKey) (*ecdsaSigner, error) {
	if pk == nil {
		return nil, errors.New("signer: private key cannot be nil")
	}
	return &ecdsaSigner{privKey: pk}, nil
}

// Address returns the address associated with the ECDSA signer.
func (s *ecdsaSigner) Address() common.Address {
	return crypto.PubkeyToAddress(s.privKey.PublicKey)
}

// SignerFn returns a signer function using the ECDSA private key and chain ID.
func (s *ecdsaSigner) SignerFn(chainID *big.Int) bind.SignerFn {
	return opcrypto.PrivateKeySignerFn(s.privKey, chainID)
}

// SignData signs the given data using the ECDSA private key.
// It applies the EIP-191 prefix ("\x19Ethereum Signed Message:\n") to prevent signing oracle attacks.
func (s *ecdsaSigner) SignData(data []byte) ([]byte, error) {
	// accounts.TextHash automatically handles the EIP-191 prefixing and hashing.
	hash := accounts.TextHash(data)

	sig, err := crypto.Sign(hash, s.privKey)
	if err != nil {
		return nil, err
	}

	// Adjust the recovery ID to match Ethereum 'v' value (27 or 28)
	sig[crypto.RecoveryIDOffset] += 27
	return sig, nil
}
