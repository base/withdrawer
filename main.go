package main

import (
	"context"
	"flag"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-node/bindings"
	bindingspreview "github.com/ethereum-optimism/optimism/op-node/bindings/preview"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/base/withdrawer/signer"
	"github.com/base/withdrawer/withdraw"
)

type network struct {
	l2RPC              string
	portalAddress      string
	l2OOAddress        string
	disputeGameFactory string
	faultProofs        bool
}

// GasConfig holds gas-related configuration for transactions
type GasConfig struct {
	GasLimit       uint64   // Override automatic gas estimation
	GasPrice       *big.Int // Legacy transaction gas price
	MaxFeePerGas   *big.Int // EIP-1559 max fee per gas
	MaxPriorityFee *big.Int // EIP-1559 max priority fee
	GasMultiplier  float64  // Multiplier for estimated gas (default 1.0)
	MaxGasPrice    *big.Int // Safety cap on gas price
}

var networks = map[string]network{
	"base-mainnet": {
		l2RPC:              "https://mainnet.base.org",
		portalAddress:      "0x49048044D57e1C92A77f79988d21Fa8fAF74E97e",
		l2OOAddress:        "0x0000000000000000000000000000000000000000",
		disputeGameFactory: "0x43edB88C4B80fDD2AdFF2412A7BebF9dF42cB40e",
		faultProofs:        true,
	},
	"base-sepolia": {
		l2RPC:              "https://sepolia.base.org",
		portalAddress:      "0x49f53e41452C74589E85cA1677426Ba426459e85",
		l2OOAddress:        "0x0000000000000000000000000000000000000000",
		disputeGameFactory: "0xd6E6dBf4F7EA0ac412fD8b65ED297e64BB7a06E1",
		faultProofs:        true,
	},
	"op-mainnet": {
		l2RPC:              "https://mainnet.optimism.io",
		portalAddress:      "0xbEb5Fc579115071764c7423A4f12eDde41f106Ed",
		l2OOAddress:        "0x0000000000000000000000000000000000000000",
		disputeGameFactory: "0xe5965Ab5962eDc7477C8520243A95517CD252fA9",
		faultProofs:        true,
	},
	"op-sepolia": {
		l2RPC:              "https://sepolia.optimism.io",
		portalAddress:      "0x16Fc5058F25648194471939df75CF27A2fdC48BC",
		l2OOAddress:        "0x0000000000000000000000000000000000000000",
		disputeGameFactory: "0x05F9613aDB30026FFd634f38e5C4dFd30a197Fa1",
		faultProofs:        true,
	},
}

func main() {
	var networkKeys []string
	for n := range networks {
		networkKeys = append(networkKeys, n)
	}

	var rpcFlag string
	var networkFlag string
	var l2RpcFlag string
	var faultProofs bool
	var portalAddress string
	var l2OOAddress string
	var dgfAddress string
	var withdrawalFlag string
	var privateKey string
	var ledger bool
	var mnemonic string
	var hdPath string

	// Gas configuration flags
	var gasLimit uint64
	var gasPrice string
	var maxFeePerGas string
	var maxPriorityFee string
	var gasMultiplier float64
	var maxGasPrice string

	flag.StringVar(&rpcFlag, "rpc", "", "Ethereum L1 RPC url")
	flag.StringVar(&networkFlag, "network", "base-mainnet", fmt.Sprintf("op-stack network to withdraw.go from (one of: %s)", strings.Join(networkKeys, ", ")))
	flag.StringVar(&l2RpcFlag, "l2-rpc", "", "Custom network L2 RPC url")
	flag.BoolVar(&faultProofs, "fault-proofs", false, "Use fault proofs")
	flag.StringVar(&portalAddress, "portal-address", "", "Custom network OptimismPortal address")
	flag.StringVar(&l2OOAddress, "l2oo-address", "", "Custom network L2OutputOracle address")
	flag.StringVar(&dgfAddress, "dgf-address", "", "Custom network DisputeGameFactory address")
	flag.StringVar(&withdrawalFlag, "withdrawal", "", "TX hash of the L2 withdrawal transaction")
	flag.StringVar(&privateKey, "private-key", "", "Private key to use for signing transactions")
	flag.BoolVar(&ledger, "ledger", false, "Use ledger device for signing transactions")
	flag.StringVar(&mnemonic, "mnemonic", "", "Mnemonic to use for signing transactions")
	flag.StringVar(&hdPath, "hd-path", "m/44'/60'/0'/0/0", "Hierarchical deterministic derivation path for mnemonic or ledger")

	// Gas configuration flags
	flag.Uint64Var(&gasLimit, "gas-limit", 0, "Gas limit for transactions (overrides automatic estimation)")
	flag.StringVar(&gasPrice, "gas-price", "", "Gas price in wei for legacy transactions")
	flag.StringVar(&maxFeePerGas, "max-fee-per-gas", "", "Maximum fee per gas in wei for EIP-1559 transactions")
	flag.StringVar(&maxPriorityFee, "max-priority-fee", "", "Maximum priority fee per gas in wei for EIP-1559 transactions")
	flag.Float64Var(&gasMultiplier, "gas-multiplier", 1.0, "Multiplier for estimated gas limit (default 1.0)")
	flag.StringVar(&maxGasPrice, "max-gas-price", "", "Maximum gas price cap in wei (safety limit)")

	flag.Parse()

	log.SetDefault(oplog.NewLogger(os.Stderr, oplog.DefaultCLIConfig()))

	n, ok := networks[networkFlag]
	if !ok {
		log.Crit("Unknown network", "network", networkFlag)
	}

	// check for non-compatible networks with given flags
	if faultProofs {
		if n.faultProofs == false {
			log.Crit("Fault proofs are not supported on this network")
		}
	} else {
		if n.faultProofs == true {
			log.Crit("Fault proofs are required on this network, please provide the --fault-proofs flag")
		}
	}

	// check for non-empty flags for non-fault proof networks
	if !faultProofs && (l2RpcFlag != "" || portalAddress != "" || l2OOAddress != "") {
		if l2RpcFlag == "" {
			log.Crit("Missing --l2-rpc flag")
		}
		if portalAddress == "" {
			log.Crit("Missing --portal-address flag")
		}
		if l2OOAddress == "" {
			log.Crit("Missing --l2oo-address flag")
		}
		n = network{
			l2RPC:         l2RpcFlag,
			portalAddress: portalAddress,
			l2OOAddress:   l2OOAddress,
			faultProofs:   faultProofs,
		}
	}

	// check for non-empty flags for fault proof networks
	if faultProofs && (l2RpcFlag != "" || dgfAddress != "" || portalAddress != "") {
		if l2RpcFlag == "" {
			log.Crit("Missing --l2-rpc flag")
		}
		if dgfAddress == "" {
			log.Crit("Missing --dgf-address flag")
		}
		if portalAddress == "" {
			log.Crit("Missing --portal-address flag")
		}
		n = network{
			l2RPC:              l2RpcFlag,
			portalAddress:      portalAddress,
			disputeGameFactory: dgfAddress,
			faultProofs:        faultProofs,
		}
	}

	if rpcFlag == "" {
		log.Crit("Missing --rpc flag")
	}

	if withdrawalFlag == "" {
		log.Crit("Missing --withdrawal flag")
	}
	withdrawal := common.HexToHash(withdrawalFlag)

	options := 0
	if privateKey != "" {
		options++
	}
	if ledger {
		options++
	}
	if mnemonic != "" {
		options++
	}
	if options != 1 {
		log.Crit("One (and only one) of --private-key, --ledger, --mnemonic must be set")
	}

	// Parse and validate gas configuration
	gasConfig := GasConfig{
		GasLimit:      gasLimit,
		GasMultiplier: gasMultiplier,
	}

	// Parse gas price (legacy transactions)
	if gasPrice != "" {
		gasPriceBig, ok := new(big.Int).SetString(gasPrice, 10)
		if !ok {
			log.Crit("Invalid --gas-price value", "value", gasPrice)
		}
		gasConfig.GasPrice = gasPriceBig
	}

	// Parse max fee per gas (EIP-1559)
	if maxFeePerGas != "" {
		maxFeeBig, ok := new(big.Int).SetString(maxFeePerGas, 10)
		if !ok {
			log.Crit("Invalid --max-fee-per-gas value", "value", maxFeePerGas)
		}
		gasConfig.MaxFeePerGas = maxFeeBig
	}

	// Parse max priority fee (EIP-1559)
	if maxPriorityFee != "" {
		maxPriorityBig, ok := new(big.Int).SetString(maxPriorityFee, 10)
		if !ok {
			log.Crit("Invalid --max-priority-fee value", "value", maxPriorityFee)
		}
		gasConfig.MaxPriorityFee = maxPriorityBig
	}

	// Parse max gas price (safety cap)
	if maxGasPrice != "" {
		maxGasPriceBig, ok := new(big.Int).SetString(maxGasPrice, 10)
		if !ok {
			log.Crit("Invalid --max-gas-price value", "value", maxGasPrice)
		}
		gasConfig.MaxGasPrice = maxGasPriceBig
	}

	// Validate gas configuration
	if gasConfig.GasPrice != nil && (gasConfig.MaxFeePerGas != nil || gasConfig.MaxPriorityFee != nil) {
		log.Crit("Cannot use --gas-price with EIP-1559 flags (--max-fee-per-gas, --max-priority-fee)")
	}

	// If one EIP-1559 flag is set, both should be set for clarity
	if (gasConfig.MaxFeePerGas != nil) != (gasConfig.MaxPriorityFee != nil) {
		log.Crit("Both --max-fee-per-gas and --max-priority-fee must be set for EIP-1559 transactions")
	}

	// Validate gas multiplier
	if gasConfig.GasMultiplier < 1.0 {
		log.Crit("--gas-multiplier must be >= 1.0", "value", gasConfig.GasMultiplier)
	}

	// Warn if gas multiplier is set but explicit gas limit is also provided
	if gasConfig.GasMultiplier > 1.0 && gasConfig.GasLimit > 0 {
		log.Warn("--gas-multiplier is ignored when --gas-limit is explicitly set", "gas-multiplier", gasConfig.GasMultiplier, "gas-limit", gasConfig.GasLimit)
	}

	// Validate max gas price cap against configured gas prices
	if gasConfig.MaxGasPrice != nil {
		if gasConfig.GasPrice != nil && gasConfig.GasPrice.Cmp(gasConfig.MaxGasPrice) > 0 {
			log.Crit("--gas-price exceeds --max-gas-price safety cap", "gas-price", gasConfig.GasPrice, "max-gas-price", gasConfig.MaxGasPrice)
		}
		if gasConfig.MaxFeePerGas != nil && gasConfig.MaxFeePerGas.Cmp(gasConfig.MaxGasPrice) > 0 {
			log.Crit("--max-fee-per-gas exceeds --max-gas-price safety cap", "max-fee-per-gas", gasConfig.MaxFeePerGas, "max-gas-price", gasConfig.MaxGasPrice)
		}
	}

	// instantiate shared variables
	s, err := signer.CreateSigner(privateKey, mnemonic, hdPath)
	if err != nil {
		log.Crit("Error creating signer", "error", err)
	}

	withdrawer, err := CreateWithdrawHelper(rpcFlag, withdrawal, n, s, gasConfig)
	if err != nil {
		log.Crit("Error creating withdrawer", "error", err)
	}

	// handle withdrawals with or without the fault proofs withdrawer
	isFinalized, err := withdrawer.IsProofFinalized()
	if err != nil {
		log.Crit("Error querying withdrawal finalization status", "error", err)
	}
	if isFinalized {
		fmt.Println("Withdrawal already finalized")
		return
	}

	// TODO: Add functionality to generate output root proposal and prove to that proposal for FPs
	err = withdrawer.CheckIfProvable()
	if err != nil {
		log.Crit("Withdrawal is not provable", "error", err)
	}

	proofTime, err := withdrawer.GetProvenWithdrawalTime()
	if err != nil {
		log.Crit("Error querying withdrawal proof", "error", err)
	}

	if proofTime == 0 {
		err = withdrawer.ProveWithdrawal()
		if err != nil {
			log.Crit("Error proving withdrawal", "error", err)
		}

		if faultProofs {
			fmt.Println("The withdrawal has been successfully proven, finalization of the withdrawal can be done once the dispute game has finished and the finalization period has elapsed")
		} else {
			fmt.Println("The withdrawal has been successfully proven, finalization of the withdrawal can be done once the finalization period has elapsed")
		}
		return
	}

	// TODO: Add edge-case handling for FPs if a withdrawal needs to be re-proven due to blacklisted / failed dispute game resolution
	err = withdrawer.FinalizeWithdrawal()
	if err != nil {
		log.Crit("Error completing withdrawal", "error", err)
	}
}

func CreateWithdrawHelper(l1Rpc string, withdrawal common.Hash, n network, s signer.Signer, gasConfig GasConfig) (withdraw.WithdrawHelper, error) {
	ctx := context.Background()

	l1Client, err := ethclient.DialContext(ctx, l1Rpc)
	if err != nil {
		return nil, fmt.Errorf("Error dialing L1 client: %w", err)
	}

	l1ChainID, err := l1Client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error querying chain ID: %w", err)
	}

	l1Nonce, err := l1Client.PendingNonceAt(ctx, s.Address())
	if err != nil {
		return nil, fmt.Errorf("Error querying nonce: %w", err)
	}

	l1opts := &bind.TransactOpts{
		From:    s.Address(),
		Signer:  s.SignerFn(l1ChainID),
		Context: ctx,
		Nonce:   big.NewInt(int64(l1Nonce)),
	}

	// Apply gas configuration to TransactOpts
	if gasConfig.GasLimit > 0 {
		l1opts.GasLimit = gasConfig.GasLimit
		log.Info("Using custom gas limit", "gas-limit", gasConfig.GasLimit)
	}

	// Log gas multiplier if set (actual application happens in withdraw functions)
	if gasConfig.GasMultiplier > 1.0 && gasConfig.GasLimit == 0 {
		log.Info("Using gas multiplier", "multiplier", gasConfig.GasMultiplier)
	}

	// Apply legacy gas price or EIP-1559 pricing
	if gasConfig.GasPrice != nil {
		l1opts.GasPrice = gasConfig.GasPrice
		log.Info("Using legacy gas price", "gas-price", gasConfig.GasPrice.String())
	} else if gasConfig.MaxFeePerGas != nil && gasConfig.MaxPriorityFee != nil {
		l1opts.GasFeeCap = gasConfig.MaxFeePerGas
		l1opts.GasTipCap = gasConfig.MaxPriorityFee
		log.Info("Using EIP-1559 gas pricing", "max-fee-per-gas", gasConfig.MaxFeePerGas.String(), "max-priority-fee", gasConfig.MaxPriorityFee.String())
	} else {
		// No gas price specified - will use RPC defaults
		// Log estimated gas price for visibility
		suggestedGasPrice, err := l1Client.SuggestGasPrice(ctx)
		if err != nil {
			log.Warn("Failed to get suggested gas price", "error", err)
		} else {
			log.Info("Using RPC suggested gas price", "suggested-gas-price", suggestedGasPrice.String())

			// Apply max gas price safety cap if configured
			if gasConfig.MaxGasPrice != nil && suggestedGasPrice.Cmp(gasConfig.MaxGasPrice) > 0 {
				return nil, fmt.Errorf("suggested gas price %s exceeds max gas price cap %s", suggestedGasPrice.String(), gasConfig.MaxGasPrice.String())
			}
		}
	}

	// Log max gas price safety cap if configured
	if gasConfig.MaxGasPrice != nil {
		log.Info("Max gas price safety cap enabled", "max-gas-price", gasConfig.MaxGasPrice.String())
	}

	l2Client, err := rpc.DialContext(ctx, n.l2RPC)
	if err != nil {
		return nil, fmt.Errorf("Error dialing L2 client: %w", err)
	}

	if n.faultProofs {
		portal, err := bindingspreview.NewOptimismPortal2(common.HexToAddress(n.portalAddress), l1Client)
		if err != nil {
			return nil, fmt.Errorf("Error binding OptimismPortal2 contract: %w", err)
		}

		dgf, err := bindings.NewDisputeGameFactory(common.HexToAddress(n.disputeGameFactory), l1Client)
		if err != nil {
			return nil, fmt.Errorf("Error binding DisputeGameFactory contract: %w", err)
		}

		return &withdraw.FPWithdrawer{
			Ctx:           ctx,
			L1Client:      l1Client,
			L2Client:      l2Client,
			L2TxHash:      withdrawal,
			Portal:        portal,
			Factory:       dgf,
			Opts:          l1opts,
			GasMultiplier: gasConfig.GasMultiplier,
		}, nil
	} else {
		portal, err := bindings.NewOptimismPortal(common.HexToAddress(n.portalAddress), l1Client)
		if err != nil {
			return nil, fmt.Errorf("Error binding OptimismPortal contract: %w", err)
		}

		l2oo, err := bindings.NewL2OutputOracle(common.HexToAddress(n.l2OOAddress), l1Client)
		if err != nil {
			return nil, fmt.Errorf("Error binding L2OutputOracle contract: %w", err)
		}

		return &withdraw.Withdrawer{
			Ctx:           ctx,
			L1Client:      l1Client,
			L2Client:      l2Client,
			L2TxHash:      withdrawal,
			Portal:        portal,
			Oracle:        l2oo,
			Opts:          l1opts,
			GasMultiplier: gasConfig.GasMultiplier,
		}, nil
	}
}
