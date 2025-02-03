import path from "node:path";

import * as config from "./config.js";

function updateGenesis(genesis) {
	const a = genesis["app_state"];
	a.chain_id = "native-t1";
	a.staking.params.key_rotation_fee.denom = "untiv";
	a.gov.params.voting_period = "600s"; // 10min
	a.gov.params.expedited_voting_period = "60s";
	a.gov.params.expedited_quorum = "0.51";
	a.gov.params.min_deposit[0].amount = "1000000"; // 1NTIV
	a.bank.denom_metadata = [config.ntivMetadata];

	a.slashing.params.signed_blocks_window = "10000"; // as suggested per validators, originally = 100
	a.slashing.params.min_signed_per_window = "0.1";
	a.slashing.params.slash_fraction_double_sign = "0.05";
	a.slashing.params.slash_fraction_downtime = "0.0001";

	return genesis;
}

function appConfig(cfg) {
	cfg.grpc.address = ":9090"; // localhost:
	cfg["grpc-gateway"].address = "localhost:1317";
	cfg.store["app-db-backend"] = "pebbledb";
	cfg.store.options["sc-pruning-option"]["keep-recent"] = config.blocksPerHour * 6;
	cfg.store.options["sc-pruning-option"].interval = config.blocksPerHour; // 0=disable prunning
	// cmd init sets 0.1 by default
	// cfg.server["minimum-gas-prices"] = "0.08untiv"; // NOTE: in mainnet we will use 0.08 probably

	return cfg;
}

function cometConfig(cfg) {
	// Use "tcp://127.0.0.1:26657" (default) to disable RPC access from the internet
	cfg.rpc.laddr = "tcp://0.0.0.0:26657";
	cfg.db_backend = "pebbledb";
	cfg.log_level = "*:info";
	// TODO: must be enabled after starting a chain
	// cfg.statesync.enable = true;

	// mainnet
	// at minimum, filters out weaker peers. Which in turn helps to have correct logs and a reliable connection.
	// comet docs: query the ABCI app on connecting to a new peer so the app can decide if we should keep the connection or not
	// cfg.filter_peers = true

	return cfg;
}

config.updateToml(cometConfig, config.cometCfgPath, false);
config.updateToml(appConfig, config.appCfgPath, false);
config.updateJson(updateGenesis, config.genesisPath);
