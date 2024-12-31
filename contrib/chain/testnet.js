import fs from "node:fs";
import os from "node:os";
import path from "node:path";

import toml from "smol-toml";

// 1block / 5s
const blocks_per_hour = 12 * 60;
const ntivMetadata = {
	description: "The operational token of the Native blockchain.",
	denom_units: [
		{ denom: "untiv", exponent: 0, aliases: ["microntiv"] },
		{ denom: "NTIV", exponent: 6, aliases: [] },
	],
	base: "untiv",
	display: "NTIV",
	name: "NTIV",
	symbol: "NTIV",
	uri: "",
	uri_hash: "",
};

function updateGenesis(filename) {
	const genesis = JSON.parse(fs.readFileSync(filename));

	const a = genesis["app_state"];
	a.chain_id = "native-t1";
	a.staking.params.key_rotation_fee.denom = "untiv";
	a.gov.params.voting_period = "600s"; // 10min
	a.gov.params.expedited_voting_period = "60s";
	a.gov.params.min_deposit[0].amount = "1000000"; // 1NTIV
	a.bank.denom_metadata = [ntivMetadata];

	fs.writeFileSync(filename, JSON.stringify(genesis, null, 2));
}

function mkBackup(filename) {
	const l = filename.lastIndexOf(".");
	const b = filename.slice(0, l) + "-back" + filename.slice(l);
	fs.copyFileSync(filename, b);
}

function readToml(filename) {
	let data = fs.readFileSync(filename, "utf8");
	return toml.parse(data);
}

function updateAppConfig(filename, backup = false) {
	if (backup) mkBackup(filename);

	const cfg = readToml(filename);
	cfg.grpc.address = ":9090"; // localhost:
	cfg["grpc-gateway"].address = ":1317";
	// TODO: https://github.com/cosmos/cosmos-sdk/issues/23133
	// cfg.store["app-db-backend"] = "pebbledb";
	// cfg.store["app-db-backend"] = "rocks";
	cfg.store.options["sc-pruning-option"]["keep-recent"] = blocks_per_hour * 2;
	cfg.store.options["sc-pruning-option"].interval = blocks_per_hour * 3; // 0=disable prunning
	// not needed: cfg.server["minimum-gas-prices"] = "0.08untiv";

	fs.writeFileSync(filename, toml.stringify(cfg));
}

function updateCometConfig(filename, backup = true) {
	if (backup) mkBackup(filename);

	const cfg = readToml(filename);
	// Use "tcp://127.0.0.1:26657" (default) to disable RPC access from the internet
	cfg.rpc.laddr = "tcp://0.0.0.0:26657";
	// TODO: must be enabled after starting a chain cfg.statesync.enable = true;
	cfg.db_backend = "pebbledb";
	cfg.log_level = "*:info";

	fs.writeFileSync(filename, toml.stringify(cfg));
}

const cfgPath = path.join(os.homedir(), ".gonative", "config");
const cometCfgPath = path.join(cfgPath, "config.toml");
const appCfgPath = path.join(cfgPath, "app.toml");

updateGenesis(path.join(cfgPath, "genesis.json"));
updateCometConfig(cometCfgPath, false);
updateAppConfig(appCfgPath, true);
