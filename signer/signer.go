package signer

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/usbwallet"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// Signer defines the interface for interacting with different types of signers.
type Signer interface {
	Address() common.Address                 // Address returns the Ethereum address associated with the signer.
	SignerFn(chainID *big.Int) bind.SignerFn // SignerFn returns a signer function used for transaction signing.
	SignData([]byte) ([]byte, error)         // SignData signs the given data using the signer's private key.
}

// CreateSigner creates a signer based on the provided private key, mnemonic, or hardware wallet.
func CreateSigner(privateKey, mnemonic, hdPath string) (Signer, error) {
	// 1. Handle Private Key String
	if privateKey != "" {
		key, err := crypto.HexToECDSA(privateKey)
		if err != nil {
			return nil, fmt.Errorf("error parsing private key: %w", err)
		}
		// FIX: Use the constructor to ensure validation and safety
		return NewECDSASigner(key)
	}

	// Parse the derivation path once to avoid repetition
	path, err := accounts.ParseDerivationPath(hdPath)
	if err != nil {
		return nil, fmt.Errorf("invalid HD path: %w", err)
	}

	// 2. Handle Mnemonic
	if mnemonic != "" {
		// Note: Ensure derivePrivateKeyFromMnemonic is defined in your package
		key, err := derivePrivateKeyFromMnemonic(mnemonic, path)
		if err != nil {
			return nil, fmt.Errorf("error deriving key from mnemonic: %w", err)
		}
		// FIX: Use the constructor
		return NewECDSASigner(key)
	}

	// 3. Handle Hardware Wallet (Ledger)
	// FIX: Ledger detection is rarely instant. We must wait/scan briefly.
	ledgerHub, err := usbwallet.NewLedgerHub()
	if err != nil {
		return nil, fmt.Errorf("error starting Ledger hub: %w", err)
	}

	// Retry loop to allow the USB device to be detected
	var wallets []accounts.Wallet
	for i := 0; i < 3; i++ {
		wallets = ledgerHub.Wallets()
		if len(wallets) > 0 {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}

	if len(wallets) == 0 {
		return nil, errors.New("no Ledger device found, please connect and unlock your Ledger")
	} else if len(wallets) > 1 {
		return nil, errors.New("multiple Ledger devices found, please use only one at a time")
	}

	wallet := wallets[0]
	if err := wallet.Open(""); err != nil {
		return nil, fmt.Errorf("error opening Ledger: %w", err)
	}

	// Derive the specific account on the Ledger
	account, err := wallet.Derive(path, true)
	if err != nil {
		return nil, fmt.Errorf("error deriving Ledger account (check if Ethereum app is open): %w", err)
	}

	return &walletSigner{
		wallet:  wallet,
		account: account,
	}, nil
}

// walletSigner implements the Signer interface for hardware wallets.
type walletSigner struct {
	wallet  accounts.Wallet
	account accounts.Account
}

func (s *walletSigner) Address() common.Address {
	return s.account.Address
}

func (s *walletSigner) SignerFn(chainID *big.Int) bind.SignerFn {
	return func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
		if address != s.account.Address {
			return nil, errors.New("not authorized to sign for this address")
		}
		return s.wallet.SignTx(s.account, tx, chainID)
	}
}

func (s *walletSigner) SignData(data []byte) ([]byte, error) {
	// accounts.Wallet.SignText automatically applies the standard EIP-191 prefix
	// ("\x19Ethereum Signed Message:\n" + len(message) + message).
	return s.wallet.SignText(s.account, data)
}
