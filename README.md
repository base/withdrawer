# withdrawer

Golang utility for proving and finalizing ETH withdrawals from OP Stack chains.

<!-- Badge row 1 - status -->

[![GitHub contributors](https://img.shields.io/github/contributors/base-org/withdrawer)](https://github.com/base-org/withdrawer/graphs/contributors)
[![GitHub commit activity](https://img.shields.io/github/commit-activity/w/base-org/withdrawer)](https://github.com/base-org/withdrawer/graphs/contributors)
[![GitHub Stars](https://img.shields.io/github/stars/base-org/withdrawer.svg)](https://github.com/base-org/withdrawer/stargazers)
![GitHub repo size](https://img.shields.io/github/repo-size/base-org/withdrawer)
[![GitHub](https://img.shields.io/github/license/base-org/withdrawer?color=blue)](https://github.com/base-org/withdrawer/blob/main/LICENSE)

<!-- Badge row 2 - links and profiles -->

[![Website base.org](https://img.shields.io/website-up-down-green-red/https/base.org.svg)](https://base.org)
[![Blog](https://img.shields.io/badge/blog-up-green)](https://base.mirror.xyz/)
[![Docs](https://img.shields.io/badge/docs-up-green)](https://docs.base.org/)
[![Discord](https://img.shields.io/discord/1067165013397213286?label=discord)](https://base.org/discord)
[![Twitter BuildOnBase](https://img.shields.io/twitter/follow/BuildOnBase?style=social)](https://x.com/BuildOnBase)

<!-- Badge row 3 - detailed status -->

[![GitHub pull requests by-label](https://img.shields.io/github/issues-pr-raw/base-org/withdrawer)](https://github.com/base-org/withdrawer/pulls)
[![GitHub Issues](https://img.shields.io/github/issues-raw/base-org/withdrawer.svg)](https://github.com/base-org/withdrawer/issues)

## Installation

```bash
git clone https://github.com/base-org/withdrawer.git
cd withdrawer
go install .
```

## Usage

> **CAUTION:**
> Do not send ERC-20 or other tokens to the L2StandardBridge. Only native ETH is supported.

### Without Fault Proofs

#### Step 1

Initiate a withdrawal on L2 by sending ETH to the `L2StandardBridge` contract at `0x4200000000000000000000000000000000000010`, and note the transaction hash.

Example on Base Sepolia: [0x5e47346867cf87d8e8c82cae1d30a94b8d5587dc9d354aef5c5a7b4c84ad9463](https://sepolia.basescan.org/tx/0x5e47346867cf87d8e8c82cae1d30a94b8d5587dc9d354aef5c5a7b4c84ad9463).

> **NOTE:**
> Users are required to wait for a period of seven days when moving assets out of Base mainnet into the Ethereum mainnet. This period is called the Challenge Period and helps secure the assets stored on Base mainnet.

#### Step 2

Prove your withdrawal:

```bash
withdrawer --network base-mainnet --withdrawal <withdrawal tx hash> --rpc <L1 RPC URL> --private-key <L1 private key>
```

or use a ledger:

```bash
withdrawer --network base-mainnet --withdrawal <withdrawal tx hash> --rpc <L1 RPC URL> --ledger
```

Example output:

```text
Proved withdrawal for 0xc4055dcb2e4647c37166caba8c7392625c2b62f9117a8bc4d96270da24b38f13: 0x6b6d1cc45b6601a30646847f638847feb629221ee71bbe6a3de7e6d0fbfe8fad
waiting for tx confirmation
0x6b6d1cc45b6601a30646847f638847feb629221ee71bbe6a3de7e6d0fbfe8fad confirmed
```

_Note: This can be called from any L1 address. It does not have to be the same address that initiated the withdrawal on the L2._

#### Step 3

After the finalization period, finalize your withdrawal (same command as above):

```bash
withdrawer --network base-mainnet --withdrawal <withdrawal tx hash> --rpc <L1 RPC URL> --private-key <L1 private key>
```

Example output:

```text
Completed withdrawal for 0xc4055dcb2e4647c37166caba8c7392625c2b62f9117a8bc4d96270da24b38f13: 0x1c457f1992f48f1f959ceaee5b3c7e699a26f6f05d93997d49dafe703fd66dea
waiting for tx confirmation
0x1c457f1992f48f1f959ceaee5b3c7e699a26f6f05d93997d49dafe703fd66dea confirmed
```

### With Fault Proofs

> **NOTE:**
> With the recent fault proofs upgrade for Base on Sepolia testnet, withdrawals require a seven-day waiting period. This mirrors the Challenge Period on Base mainnet. Withdrawals must also be finalized against dispute games that resolve in favor of the output root claim. If the dispute game is blacklisted or resolves against the root claim, the withdrawal must be re-proven.

#### Step 1

Initiate a withdrawal on L2 by sending ETH to the `L2StandardBridge` contract at `0x4200000000000000000000000000000000000010`, and note the transaction hash.

Example on Base Sepolia: [0x5e47346867cf87d8e8c82cae1d30a94b8d5587dc9d354aef5c5a7b4c84ad9463](https://sepolia.basescan.org/tx/0x5e47346867cf87d8e8c82cae1d30a94b8d5587dc9d354aef5c5a7b4c84ad9463).

#### Step 2

Prove your withdrawal:

```bash
withdrawer --network base-mainnet --withdrawal <withdrawal tx hash> --rpc <L1 RPC URL> --private-key <L1 private key> --fault-proofs
```

or use a ledger:

```bash
withdrawer --network base-mainnet --withdrawal <withdrawal tx hash> --rpc <L1 RPC URL> --ledger --fault-proofs
```

Example output:

```text
Proved withdrawal for 0xc4055dcb2e4647c37166caba8c7392625c2b62f9117a8bc4d96270da24b38f13: 0x6b6d1cc45b6601a30646847f638847feb629221ee71bbe6a3de7e6d0fbfe8fad
waiting for tx confirmation
0x6b6d1cc45b6601a30646847f638847feb629221ee71bbe6a3de7e6d0fbfe8fad confirmed
```

#### Step 3

> **IMPORTANT:**
> Unlike non-fault proof withdrawals, you MUST use the same address that proved the withdrawal to finalize it.

After the dispute game resolves in favor of the root claim and the finalization period elapses, finalize your withdrawal:

```bash
withdrawer --network base-mainnet --withdrawal <withdrawal tx hash> --rpc <L1 RPC URL> --private-key <L1 private key> --fault-proofs
```

Example output:

```text
Completed withdrawal for 0xc4055dcb2e4647c37166caba8c7392625c2b62f9117a8bc4d96270da24b38f13: 0x1c457f1992f48f1f959ceaee5b3c7e699a26f6f05d93997d49dafe703fd66dea
waiting for tx confirmation
0x1c457f1992f48f1f959ceaee5b3c7e699a26f6f05d93997d49dafe703fd66dea confirmed
```

## Flags

```bash
Usage of withdrawer:
    -rpc string
        Ethereum L1 RPC URL
    -network string
        OP Stack network to withdraw from (e.g., base-mainnet, base-sepolia, op-mainnet, op-sepolia) (default: "base-mainnet")
    -withdrawal string
        TX hash of the L2 withdrawal transaction
    -fault-proofs
        Use fault proof withdrawal flow (only for networks that support fault proofs)
    -private-key string
        Private key for signing transactions
    -mnemonic string
        Mnemonic for signing transactions
    -ledger
        Use a Ledger device for signing transactions
    -hd-path string
        HD derivation path for mnemonic or Ledger (default: "m/44'/60'/0'/0/0")
    -l2-rpc string
        Custom L2 RPC URL
    -l2oo-address string
        Custom L2OutputOracle address
    -portal-address string
        Custom OptimismPortal address
    -dfg-address string
        Custom DisputeGameFactory address (only for networks that support fault proofs)
