'use strict';

require('./wasm_exec.js');
var fs = require('fs');
var path = require('path');

const CWD = process.cwd();
const EXCLUDES = [
  "nodot.test",
  "remote.test",
  "integration.test"
];

async function main() {
  for (let i = 2; i < process.argv.length; i++) {
    const dirname = path.dirname(path.join(CWD, process.argv[i]))
    const basename = path.basename(process.argv[i])

    // Skip tests that cannot be executed.
    if (EXCLUDES.includes(basename)) {
      continue;
    }

    // Exit(1) when tests failed.
    let go = new Go();
    go.exit = (code) => {
      if (code !== 0) {
        process.exit(code);
      }
    };

    // Use `chdir` since we are using files based on relative paths in tests.
    process.chdir(dirname);
    const mod = await WebAssembly.compile(fs.readFileSync(basename));
    let inst = await WebAssembly.instantiate(mod, go.importObject);

    await go.run(inst);
  }
}

main().catch(err => {
  console.error(err);
  process.exit(1);
});
