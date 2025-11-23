package signer

import (
	"crypto/ecdsa"
	"math/big"

	opcrypto "github.com/ethereum-optimism/optimism/op-service/crypto"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Signer defines the interface required for signing Ethereum transactions and messages.
// This abstraction allows for future support of different signing methods (e.g., Ledger, Mnemonic).
type Signer interface {
	// Address returns the public Ethereum address of the signer.
	Address() common.Address
	// SignerFn returns a bind.SignerFn suitable for use with go-ethereum's ABIs.
	SignerFn(chainID *big.Int) bind.SignerFn
	// SignHash signs a Keccak256 hash digest (32 bytes) and returns the R, S, V signature.
	SignHash(hash []byte) ([]byte, error)
}

// ecdsaSigner implements the Signer interface using a standard ECDSA private key.
type ecdsaSigner struct {
	// The underlying ECDSA private key.
	*ecdsa.PrivateKey
}

// NewECDsaSigner creates a new ecdsaSigner instance from an ECDSA private key.
func NewECDsaSigner(pk *ecdsa.PrivateKey) Signer {
	return &ecdsaSigner{pk}
}

// Address returns the address associated with the ECDSA signer.
func (s *ecdsaSigner) Address() common.Address {
	return crypto.PubkeyToAddress(s.PublicKey)
}

// SignerFn returns a signer function using the ECDSA private key and chain ID.
// It leverages the op-service implementation for standard transaction signing.
func (s *ecdsaSigner) SignerFn(chainID *big.Int) bind.SignerFn {
	// Note: Using op-crypto's helper for robust transaction signing logic.
	return opcrypto.PrivateKeySignerFn(s.PrivateKey, chainID)
}

// SignHash signs the given 32-byte hash digest using the ECDSA private key.
// It applies the standard Ethereum recovery ID (v) adjustment (v = v + 27).
func (s *ecdsaSigner) SignHash(hash []byte) ([]byte, error) {
	// Sanity check: ensure we are signing a 32-byte hash digest.
	if len(hash) != 32 {
		return nil, crypto.ErrInvalidHash
	}

	sig, err := crypto.Sign(hash, s.PrivateKey)
	if err != nil {
		return nil, err
	}

	// Adjust the recovery ID (V) from 0/1 to 27/28 for Ethereum compatibility (Geth standard).
	sig[crypto.RecoveryIDOffset] += 27
	return sig, nil
}
