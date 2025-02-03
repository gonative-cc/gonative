import fs from "node:fs";
import os from "node:os";
import path from "node:path";

import toml from "smol-toml";

export const cfgPath = path.join(os.homedir(), ".gonative", "config");
export const cometCfgPath = path.join(cfgPath, "config.toml");
export const appCfgPath = path.join(cfgPath, "app.toml");
export const genesisPath = path.join(cfgPath, "genesis.json");

// 1block / 5s
export const blocksPerHour = 12 * 60;
export const ntivMetadata = {
	description: "The operational token of the Native blockchain.",
	denom_units: [
		{ denom: "untiv", exponent: 0, aliases: ["microntiv"] },
		{ denom: "NTIV", exponent: 6, aliases: [] },
	],
	base: "untiv",
	display: "NTIV",
	name: "Native",
	symbol: "NTIV",
	uri: "",
	uri_hash: "",
};

function mkBackup(filename) {
	const l = filename.lastIndexOf(".");
	const b = filename.slice(0, l) + "-back" + filename.slice(l);
	fs.copyFileSync(filename, b);
}

function readToml(filename) {
	let data = fs.readFileSync(filename, "utf8");
	return toml.parse(data);
}

export function updateToml(processor, filename, backup = true) {
	if (backup) mkBackup(filename);

	let cfg = readToml(filename);
	cfg = processor(cfg);
	fs.writeFileSync(filename, toml.stringify(cfg));
}

export function updateJson(processor, filename, backup = true) {
	if (backup) mkBackup(filename);

	let genesis = JSON.parse(fs.readFileSync(filename));
	genesis = processor(genesis);
	fs.writeFileSync(filename, JSON.stringify(genesis, null, 2));
}
