<!-- markdownlint-disable MD041 -->
<!-- markdownlint-disable MD013 -->

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue?style=flat-square&logo=go)](https://godoc.org/https://github.com/gonative-cc/gonative)
[![Go Report Card](https://goreportcard.com/badge/github.com/gonative-cc/gonative?style=flat-square)](https://goreportcard.com/report/https://github.com/gonative-cc/gonative)
[![Version](https://img.shields.io/github/tag/gonative-cc/gonative.svg?style=flat-square)](https://github.com/gonative-cc/gonative/releases/latest)
[![License: Apache-2.0](https://img.shields.io/github/license/gonative-cc/gonative.svg?style=flat-square)](https://github.com/gonative-cc/gonative/blob/main/LICENSE)

# Go Native

A Go lang implementation of the Native - a secure and decentralized Interoperability and Application Platform for Bitcoin based on the Zero Trust Architecture.

Native is transforming Bitcoin into a programmable, decentralized finance powerhouse without sacrificing its core values. Leveraging the groundbreaking Zero Trust Architecture, Native brings modular interoperability to Bitcoin, allowing dApps to securely tap into Bitcoin's vast liquidity and yield opportunities.

- [Website](https://www.gonative.cc/)
- [x.com/NativeNetwork](https://x.com/NativeNetwork)
- [Discord](https://discord.com/invite/gonative)

## Release Compatibility Matrix

| Version | Testnet | Mainnet | Cosmos SDK | IBC  | libwasmvm |
| :-----: | :-----: | :-----: | :--------: | :--: | :-------: |
|  todo   |    ✓    |    ✗    |  v0.52.x   | TODO |     -     |

## Contributing

Participating in open source is often a highly collaborative experience. We’re encouraged to create in public view, and we’re incentivized to welcome contributions of all kinds from people around the world.

Check out [contributing repo](https://github.com/gonative-cc/contributig) for our guidelines & policies for how to contribute. Note: we require DCO! Thank you to all those who have contributed!

After cloning the repository, make sure to run `make setup-hooks`.

### Security

Check out [SECURITY.md](./SECURITY.md) for security concerns.

## Setup

### Build

```shell
make build
```

- [libwasmvm notes](https://github.com/gonative-cc/network-docs/blob/master/validator.md#libwasmvm)

### Recommended Database Backend

We recommend to use Pebble. Note: RocksDB is not supported.
Make sure you have it set in the config files:

```bash
# app.toml / base configuration options
app-db-backend = "pebbledb"

# config.toml / base configuration options
db_backend = "pebbledb"
```

## Validators

Please follow [network documentation](https://github.com/gonative-cc/network-docs) to configure and setup your validator node.
