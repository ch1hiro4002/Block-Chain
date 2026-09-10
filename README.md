# Block-Chain

A small, educational blockchain node written in Go. The project is a learning-focused implementation of common blockchain concepts: signed transactions, a transaction mempool, block construction and validation, peer-to-peer-style message broadcasting, and a tiny stack-based virtual machine with in-memory key-value state.

中文文档：[README.zh-CN.md](docs/README.zh-CN.md)

> This project is not production-ready. It is intended for learning Go concurrency, interface design, and basic blockchain mechanics.

## Features

- ECDSA P-256 key generation, transaction signing, and signature verification
- Transaction mempool with duplicate detection and a maximum size
- Linked blocks with SHA-256 hashes and block validation
- Validator loop that periodically creates new blocks
- Local in-memory transport with message broadcasting
- Simple stack VM supporting integers, byte packing, arithmetic, and state storage/retrieval
- Unit tests for core, network, and crypto packages

## Requirements

- Go 1.26+

## Quick Start

Build and run:

```bash
make build
make run
```

Run tests:

```bash
make test
```

The demo starts several in-process nodes connected through a local transport. One remote node repeatedly sends a signed transaction to a local validator node, which adds it to the mempool and periodically creates blocks.

## Project Structure

```text
.
├── core/                 # Transaction, block, blockchain, validator, VM, state, encoding
├── crypto/               # ECDSA keys, signatures, addresses
├── network/              # Transport abstraction, local transport, RPC, server, tx pool
├── types/                # Hash, address, generic list
├── testutil/             # Random test helpers
├── util/                 # Small generic utilities
├── docs/                 # Documentation translations
├── main.go               # Demo wiring
├── Makefile
└── go.mod
```

## Main Components

### Transactions

Transactions contain:

- `Data`: transaction payload, currently interpreted as VM bytecode
- `From`: sender public key
- `Signature`: signature over `Data`

Before entering the mempool, a transaction is verified and assigned a local first-seen timestamp.

### Blocks

Blocks contain a header, a list of transactions, the validator public key, and a block signature. The header includes:

- Version
- Transaction data hash
- Previous block hash
- Timestamp
- Height

### Blockchain

`core.BlockChain` stores headers in memory and validates incoming blocks before adding them. The current storage implementation is in-memory and not persisted.

### Network

`network.Transport` is an abstraction over message delivery. The current implementation is `LocalTransport`, which connects in-process nodes using channels and Go maps.

Messages are encoded with `encoding/gob` and include either a transaction or a block.

### Virtual Machine

The VM executes transaction `Data` as a simple bytecode stream and maintains a stack plus a shared in-memory state.

Current instruction set:

| Opcode | Instruction | Stack behavior |
| --- | --- | --- |
| `0x0a` | `InstrPushInt` | Push the preceding byte as `int` |
| `0x0b` | `InstrPushByte` | Push the preceding byte as `byte` |
| `0x0c` | `InstrAdd` | Pop `a`, pop `b`, push `a + b` |
| `0x0d` | `InstrSub` | Pop `a`, pop `b`, push `b - a` |
| `0x0e` | `InstrMul` | Pop `a`, pop `b`, push `a * b` |
| `0x0f` | `InstrDiv` | Pop `a`, pop `b`, push `b / a`; reject `a == 0` |
| `0x10` | `InstrPack` | Pop size, then pop that many bytes into `[]byte` |
| `0x11` | `InstrStore` | Pop key and value, write to state |
| `0x12` | `InstrGet` | Pop key, push state value |

The bytecode format is intentionally minimal. Immediates are placed before their corresponding opcodes, for example:

```text
0x03 0x0a   push int 3
0x02 0x0a   push int 2
0x0c        add -> 5
```

## Demo Flow

1. A remote transport repeatedly creates and sends a signed transaction to the local server.
2. The local server verifies the transaction, adds it to the mempool, and broadcasts it.
3. The local validator runs on a timer and creates a new signed block from pending transactions.
4. The block is validated, executed by the VM, stored in memory, and broadcast to peers.
5. Other nodes validate and add the block to their local chain.

## Limitations

- Network transport is local and in-memory; there is no real TCP or peer discovery.
- Blockchain and state are not persisted to disk.
- Consensus is simplified to a single validator; there is no fork resolution or multi-validator consensus.
- The VM is minimal and has no gas, execution limits, or comprehensive runtime safety.
- The block header does not yet include a state root or receipt root.
- This code is for learning and demonstration, not for production or financial use.

## Roadmap Ideas

- Add a real TCP or HTTP transport
- Persist blockchain and state with a database
- Add state root/receipt root verification
- Remove confirmed transactions from remote mempools
- Add gas metering and better VM error handling
- Add peer discovery and multi-validator consensus

## Testing

```bash
go test ./...
```

The test suite covers transaction signing/verification, block validation, blockchain height handling, transaction pool behavior, and VM arithmetic/state instructions.
