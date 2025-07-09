#!/usr/bin/env node

import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { spawnSync } from "node:child_process";

const __dirname = dirname(fileURLToPath(import.meta.url));

// TODO: cross-compiled binaries
const result = spawnSync(join(__dirname, "./yangly"), process.argv.slice(2), {
	shell: false,
	stdio: "inherit",
});

if (result.error) {
	throw result.error;
}

process.exitCode = result.status;
