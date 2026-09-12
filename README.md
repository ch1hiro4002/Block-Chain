# Block-Chain

A compact Go blockchain project implementing signed transactions, a transaction mempool, linked and signed blocks, simplified proof-of-stake block production, TCP-based block synchronization, and a minimal stack-based virtual machine.

中文文档：[README.zh-CN.md](docs/README.zh-CN.md)

> This project is a demo and is not intended for production use.

## What Is Implemented

- ECDSA P-256 keys, addresses, transaction signing, and block signing
- Transaction pool with duplicate detection, nonce validation, and deadlines
- SHA-256-linked blocks with proposer signatures
- A simplified PoS-style validator set with stake-weighted proposer selection
- Validator-set snapshots per block height so late nodes can validate historical blocks correctly
- TCP transport with length-prefixed messages and `gob` encoding
- Block and validator-set synchronization between peers
- A small stack VM with arithmetic, byte packing, and in-memory key-value state

## Quick Start

Build and run the demo:

```bash
make build
make run
```

Run tests:

```bash
make test
```

The demo starts three nodes: `LOCAL`, `REMOTE`, and a later-joining `LATE` node. `LATE` connects to seed nodes, synchronizes blocks and validator history, and then participates in block production.

## Project Structure

```text
.
├── core/       # transactions, blocks, blockchain, validators, VM, state, encoding
├── crypto/     # keys, signatures, addresses
├── network/    # TCP transport, RPC, server, transaction pool
├── types/      # common types such as hashes and addresses
├── docs/       # documentation
└── main.go     # demo entry point
```

## Core Components

| Component | Summary |
| --- | --- |
| Transactions | Signed payloads with nonce and deadline; currently executed as VM bytecode |
| Blocks | Headers, transaction lists, proposer address, and block signatures |
| Blockchain | In-memory chain with hash linking and block validation |
| Consensus | Simplified PoS: stake-weighted proposer selection and height-versioned validator sets |
| Network | TCP peers exchange status, blocks, transactions, and validator snapshots |
| Virtual Machine | Minimal stack machine for arithmetic, byte packing, and state storage |

## Current Limitations

- Validator membership is bootstrapped over the network rather than governed by on-chain stake transactions
- No slashing, finality, fork-choice, or advanced multi-validator consensus
- Blockchain and state are stored in memory only
- No peer discovery, authentication, or production-grade network security
- The VM has no gas metering or execution limits

## Testing

```bash
go test ./...
```
