import { build } from 'esbuild';
import crypto from 'node:crypto';
import fs from 'node:fs/promises';
import path from 'node:path';

const externalDependencies = [
  'three',
  '@senspace/plugin-core',
  '@senspace/plugin-framework',
];

const externalModuleMap = {
  three: '/plugin-host/three.module.js',
  '@senspace/plugin-core': '/plugin-host/plugin-core.js',
  '@senspace/plugin-framework': '/plugin-host/plugin-framework.js',
};

const sandboxEntry = 'runtime/sandbox-entry.js';

function toPosixPath(value) {
  return String(value).split(path.sep).join('/');
}

function normalizeDependencyName(specifier) {
  const normalized = String(specifier || '').trim();
  if (!normalized) {
    return '';
  }
  for (const [dependencyName, externalPath] of Object.entries(externalModuleMap)) {
    if (normalized === externalPath) {
      return dependencyName;
    }
  }
  if (normalized.startsWith('@')) {
    const parts = normalized.split('/');
    return parts.length >= 2 ? `${parts[0]}/${parts[1]}` : normalized;
  }
  return normalized.split('/')[0];
}

function extractNodeModulesDependency(inputPath) {
  const normalized = toPosixPath(inputPath);
  const marker = 'node_modules/';
  const index = normalized.lastIndexOf(marker);
  if (index < 0) {
    return '';
  }
  return normalizeDependencyName(normalized.slice(index + marker.length));
}

function createFactoryWrapper({ entryImportPath, pluginId }) {
  return `
import PluginEntry from ${JSON.stringify(entryImportPath)};

const defaultPluginId = ${JSON.stringify(pluginId)};

function normalizeFactory(entry) {
  const candidate =
    entry && typeof entry === 'object' && 'default' in entry
      ? entry.default
      : entry;

  if (
    candidate &&
    typeof candidate === 'object' &&
    typeof candidate.lazyCreate === 'function'
  ) {
    return {
      pluginId: candidate.pluginId || defaultPluginId,
      lazyCreate(context, options) {
        return candidate.lazyCreate(context, options);
      },
      getPluginOptionsDefinition() {
        if (typeof candidate.getPluginOptionsDefinition !== 'function') {
          return null;
        }
        return candidate.getPluginOptionsDefinition();
      },
    };
  }

  if (typeof candidate === 'function') {
    if (typeof candidate.lazyCreate === 'function') {
      return {
        pluginId: candidate.pluginId || defaultPluginId,
        lazyCreate(context, options) {
          return candidate.lazyCreate(context, options);
        },
        getPluginOptionsDefinition() {
          if (typeof candidate.getPluginOptionsDefinition !== 'function') {
            return null;
          }
          return candidate.getPluginOptionsDefinition();
        },
      };
    }

    return {
      pluginId: defaultPluginId,
      async lazyCreate(context, options) {
        return new candidate(context, options);
      },
      getPluginOptionsDefinition() {
        if (typeof candidate.getPluginOptionsDefinition !== 'function') {
          return null;
        }
        return candidate.getPluginOptionsDefinition();
      },
    };
  }

  throw new Error(
    'plugin entry must export a plugin class or plugin factory as its default export'
  );
}

const pluginFactory = normalizeFactory(PluginEntry);

export const pluginId = pluginFactory.pluginId;
export default pluginFactory;
`;
}

function collectRuntimeDependencies(metafile) {
  const externalSet = new Set();
  const bundledSet = new Set();

  for (const output of Object.values(metafile?.outputs || {})) {
    for (const dependency of output.imports || []) {
      if (!dependency.external) {
        continue;
      }
      const normalizedName = normalizeDependencyName(dependency.path);
      if (normalizedName) {
        externalSet.add(normalizedName);
      }
    }
  }

  for (const inputPath of Object.keys(metafile?.inputs || {})) {
    const dependencyName = extractNodeModulesDependency(inputPath);
    if (dependencyName && !externalSet.has(dependencyName)) {
      bundledSet.add(dependencyName);
    }
  }

  return {
    externalDependencies: Array.from(externalSet)
      .sort()
      .map((name) => ({ name, mode: 'external' })),
    bundledDependencies: Array.from(bundledSet)
      .sort()
      .map((name) => ({ name, mode: 'bundled' })),
  };
}

// 创建浏览器直载产物的 external 映射，保留旧入口过渡兼容。
function createDirectRuntimeExternalPlugin() {
  return {
    name: 'senspace-plugin-direct-externals',
    setup(pluginBuild) {
      for (const [dependencyName, externalPath] of Object.entries(
        externalModuleMap
      )) {
        pluginBuild.onResolve(
          {
            filter: new RegExp(
              `^${dependencyName.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}$`
            ),
          },
          () => ({
            path: externalPath,
            external: true,
          })
        );
      }
    },
  };
}

// 创建沙箱产物的 external 映射，让 Worker 内的受控 require 接管平台模块。
function createSandboxRuntimeExternalPlugin() {
  return {
    name: 'senspace-plugin-sandbox-externals',
    setup(pluginBuild) {
      for (const dependencyName of externalDependencies) {
        pluginBuild.onResolve(
          {
            filter: new RegExp(
              `^${dependencyName.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}$`
            ),
          },
          () => ({
            path: dependencyName,
            external: true,
          })
        );
      }
    },
  };
}

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
  const entryImportPath = (() => {
    const relativePath = toPosixPath(path.relative(srcDir, entryFile));
    return relativePath.startsWith('.') ? relativePath : `./${relativePath}`;
  })();

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
    plugins: [createDirectRuntimeExternalPlugin()],
    stdin: {
      contents: createFactoryWrapper({
        entryImportPath,
        pluginId,
      }),
      resolveDir: srcDir,
      sourcefile: '__factory_entry__.ts',
      loader: 'ts',
    },
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
  });

  const sandboxBuildResult = await build({
    plugins: [createSandboxRuntimeExternalPlugin()],
    stdin: {
      contents: createFactoryWrapper({
        entryImportPath,
        pluginId,
      }),
      resolveDir: srcDir,
      sourcefile: '__sandbox_entry__.ts',
      loader: 'ts',
    },
    outfile: path.join(outDir, sandboxEntry),
    bundle: true,
    format: 'cjs',
    platform: 'browser',
    target: ['es2022'],
    sourcemap: false,
    splitting: false,
    metafile: true,
    assetNames: 'assets/[name]-[hash]',
  });

  const runtimeEntryPath = path.join(outDir, sandboxEntry);
  const runtimeEntry = await fs.readFile(runtimeEntryPath);
  const bundleHash = `sha256:${crypto.createHash('sha256').update(runtimeEntry).digest('hex')}`;
  const integrity = `sha384-${crypto.createHash('sha384').update(runtimeEntry).digest('base64')}`;

  const runtimeDependencies = collectRuntimeDependencies(sandboxBuildResult.metafile);
  const runtimeManifest = {
    pluginId,
    version,
    releaseId,
    bundleHash,
    integrity,
    sandboxEntry,
    externalDependencies: runtimeDependencies.externalDependencies,
    bundledDependencies: runtimeDependencies.bundledDependencies,
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
