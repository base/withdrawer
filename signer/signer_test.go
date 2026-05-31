package signer

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestEcdsaSigner_Address(t *testing.T) {
	// Test that ecdsaSigner returns correct address from private key
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	signer := &ecdsaSigner{privateKey}
	addr := signer.Address()

	expectedAddr := crypto.PubkeyToAddress(privateKey.PublicKey)
	if addr != expectedAddr {
		t.Errorf("expected address %s, got %s", expectedAddr.Hex(), addr.Hex())
	}
}

func TestEcdsaSigner_SignerFn(t *testing.T) {
	// Test that SignerFn returns a valid signer function
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	signer := &ecdsaSigner{privateKey}
	signerFn := signer.SignerFn(big.NewInt(1))

	if signerFn == nil {
		t.Fatal("expected non-nil signer function")
	}

	// Verify it's a bind.SignerFn type
	var fn bind.SignerFn = signerFn
	if fn == nil {
		t.Error("signer function should be assignable to bind.SignerFn")
	}
}

func TestCreateSigner_WithPrivateKey(t *testing.T) {
	// Test CreateSigner with valid private key
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	privateKeyHex := common.Bytes2Hex(crypto.FromECDSA(privateKey))

	signer, err := CreateSigner(privateKeyHex, "", "")
	if err != nil {
		t.Fatalf("failed to create signer with private key: %v", err)
	}

	if signer == nil {
		t.Fatal("expected non-nil signer")
	}

	// Verify the signer has the correct address
	expectedAddr := crypto.PubkeyToAddress(privateKey.PublicKey)
	if signer.Address() != expectedAddr {
		t.Errorf("expected address %s, got %s", expectedAddr.Hex(), signer.Address().Hex())
	}
}

func TestCreateSigner_InvalidPrivateKey(t *testing.T) {
	// Test CreateSigner with invalid private key
	_, err := CreateSigner("invalid-key", "", "")
	if err == nil {
		t.Error("expected error for invalid private key")
	}
}

func TestCreateSigner_EmptyInputs(t *testing.T) {
	// Test CreateSigner with no inputs - should try ledger
	// This will fail since there's no ledger, but verifies the flow
	_, err := CreateSigner("", "", "")
	// We expect an error since there's no ledger connected
	if err == nil {
		t.Log("no error - either no ledger or ledger hub not available")
	}
}

func TestCreateSigner_WithMnemonic(t *testing.T) {
	// Test CreateSigner with mnemonic
	// BIP39 mnemonic for testing
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	
	signer, err := CreateSigner("", mnemonic, "m/44'/60'/0'/0/0")
	if err != nil {
		t.Fatalf("failed to create signer with mnemonic: %v", err)
	}

	if signer == nil {
		t.Fatal("expected non-nil signer")
	}

	// Verify the signer has an address
	addr := signer.Address()
	if addr == (common.Address{}) {
		t.Error("expected non-zero address")
	}
}

func TestCreateSigner_WithInvalidMnemonic(t *testing.T) {
	// Test CreateSigner with invalid mnemonic
	_, err := CreateSigner("", "invalid mnemonic that is too short", "m/44'/60'/0'/0/0")
	if err == nil {
		t.Error("expected error for invalid mnemonic")
	}
}

func TestSigner_Interface(t *testing.T) {
	// Test that our signers implement the Signer interface
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	var s Signer = &ecdsaSigner{privateKey}
	if s == nil {
		t.Fatal("ecdsaSigner should implement Signer interface")
	}

	// Verify we can call interface methods
	addr := s.Address()
	if addr == (common.Address{}) {
		t.Error("expected non-zero address from interface call")
	}

	signerFn := s.SignerFn(big.NewInt(1))
	if signerFn == nil {
		t.Error("expected non-nil signer function from interface")
	}
}

func TestSignerFn_WithDifferentChainIDs(t *testing.T) {
	// Test that SignerFn works with different chain IDs
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	signer := &ecdsaSigner{privateKey}

	chainIDs := []*big.Int{
		big.NewInt(1),      // Ethereum mainnet
		big.NewInt(8453),   // Base mainnet
		big.NewInt(11155111), // Sepolia
		big.NewInt(84532),  // Base Sepolia
	}

	for _, chainID := range chainIDs {
		signerFn := signer.SignerFn(chainID)
		if signerFn == nil {
			t.Errorf("expected non-nil signer for chain ID %s", chainID.String())
		}
	}
}

func TestEcdsaSigner_KeyPersistence(t *testing.T) {
	// Test that ecdsaSigner properly stores and uses the private key
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	signer := &ecdsaSigner{privateKey}

	// The address should be consistent
	addr1 := signer.Address()
	addr2 := signer.Address()
	if addr1 != addr2 {
		t.Error("address should be consistent across calls")
	}

	// The address should match the public key
	expectedAddr := crypto.PubkeyToAddress(privateKey.PublicKey)
	if addr1 != expectedAddr {
		t.Errorf("address mismatch: got %s, expected %s", addr1.Hex(), expectedAddr.Hex())
	}
}