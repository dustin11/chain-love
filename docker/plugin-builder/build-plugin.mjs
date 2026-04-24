import { build } from 'esbuild';
import crypto from 'node:crypto';
import fs from 'node:fs/promises';
import path from 'node:path';

const externalDependencies = [
  'three',
  '@senspace/plugin-core',
  '@senspace/plugin-framework',
];

function parseArgs(argv) {
  const result = {};
  for (let i = 0; i < argv.length; i += 1) {
    const key = argv[i];
    if (!key.startsWith('--')) {
      continue;
    }
    result[key.slice(2)] = argv[i + 1];
    i += 1;
  }
  return result;
}

async function readJSON(filePath) {
  const content = await fs.readFile(filePath, 'utf8');
  return JSON.parse(content);
}

async function fileExists(filePath) {
  try {
    await fs.access(filePath);
    return true;
  } catch {
    return false;
  }
}

async function loadManifest(srcDir) {
  const snapshotPath = path.join(srcDir, 'manifest.snapshot.json');
  if (await fileExists(snapshotPath)) {
    return readJSON(snapshotPath);
  }
  return readJSON(path.join(srcDir, 'manifest.json'));
}

async function main() {
  const args = parseArgs(process.argv.slice(2));
  const srcDir = args.src;
  const outDir = args.out;
  const pluginId = args['plugin-id'];
  const version = args.version;
  const releaseId = args['release-id'];

  if (!srcDir || !outDir || !pluginId || !version || !releaseId) {
    throw new Error('missing required args: --src --out --plugin-id --version --release-id');
  }

  const manifest = await loadManifest(srcDir);
  const entry = (manifest.entry || manifest.main || 'src/index.ts').trim();
  const entryFile = path.join(srcDir, entry);
  if (!(await fileExists(entryFile))) {
    throw new Error(`entry file not found: ${entry}`);
  }

  await fs.mkdir(path.join(outDir, 'runtime'), { recursive: true });
  await fs.copyFile(
    path.join(srcDir, 'manifest.snapshot.json'),
    path.join(outDir, 'manifest.snapshot.json'),
  ).catch(async () => {
    await fs.writeFile(
      path.join(outDir, 'manifest.snapshot.json'),
      JSON.stringify(manifest, null, 2),
      'utf8',
    );
  });

  await build({
    entryPoints: [entryFile],
    outdir: path.join(outDir, 'runtime'),
    entryNames: 'index',
    bundle: true,
    format: 'esm',
    platform: 'browser',
    target: ['es2022'],
    sourcemap: true,
    splitting: true,
    metafile: true,
    chunkNames: 'chunks/[name]-[hash]',
    assetNames: 'assets/[name]-[hash]',
    external: externalDependencies,
  });

  const runtimeEntryPath = path.join(outDir, 'runtime', 'index.js');
  const runtimeEntry = await fs.readFile(runtimeEntryPath);
  const bundleHash = `sha256:${crypto.createHash('sha256').update(runtimeEntry).digest('hex')}`;
  const integrity = `sha384-${crypto.createHash('sha384').update(runtimeEntry).digest('base64')}`;

  const runtimeManifest = {
    pluginId,
    version,
    releaseId,
    bundleHash,
    integrity,
    externalDependencies: externalDependencies.map((name) => ({
      name,
      mode: 'external',
    })),
    bundledDependencies: [],
  };
  await fs.writeFile(
    path.join(outDir, 'runtime-manifest.json'),
    JSON.stringify(runtimeManifest, null, 2),
    'utf8',
  );
}

main().catch((error) => {
  console.error(error instanceof Error ? error.message : String(error));
  process.exit(1);
});
