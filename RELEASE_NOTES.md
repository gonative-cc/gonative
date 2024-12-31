<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->
<!-- markdownlint-disable MD040 -->

# Release Notes

The Release Procedure is defined in the [contributing](https://github.com/gonative-cc/contributig) repository.

## v0.1.0

First working release for the Native testnet.
We had few problems launching and configuring the testnet. We based the code on the early version of Cosmos SDK v0.52 which experienced significant delays. Keeping up with patches and finding a right configuration of modules that were not released was breakneck 🤕! We decided to wait for the Cosmos SDK release candidate, and redo the chain.
This is the first, tested version. There are ongoing issues with the Cosmos SDK RC (notably the RocksDB and Pebble support) as well as finalizing the storage integration.

We ended up with a working configuration that we are happy to release in 2024 🎉! In the next release we will migrate first version of our dwallet module.

You can build the binary yourself by following the instructions in the README file or use the attached binary.

Instructions to join the testnet are shared in Native Discord.

[CHANGELOG](./CHANGELOG.md)
