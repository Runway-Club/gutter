// Tiny static file server for the benchmark. Serves the two built apps under
// /gutter/ and /react/ with production-like behaviour:
//   - correct MIME types (notably application/wasm),
//   - gzip when the client accepts it (Content-Encoding: gzip),
//   - cacheable responses (ETag + Cache-Control) so the warm-reload measurement
//     actually hits the browser cache instead of refetching.
import http from 'node:http';
import { promises as fs } from 'node:fs';
import { createReadStream, statSync } from 'node:fs';
import path from 'node:path';
import zlib from 'node:zlib';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

const ROOTS = {
  '/gutter': path.join(__dirname, 'gutter-app', 'dist'),
  '/gutter-tinygo': path.join(__dirname, 'gutter-app', 'dist-tinygo'),
  '/gutter-ssr': path.join(__dirname, 'gutter-app', 'dist-ssr'),
  '/gutter-ssr-tinygo': path.join(__dirname, 'gutter-app', 'dist-ssr-tinygo'),
  '/react': path.join(__dirname, 'react-app', 'dist'),
  '/compute-go': path.join(__dirname, 'compute', 'dist-go'),
  '/compute-tinygo': path.join(__dirname, 'compute', 'dist-tinygo'),
  '/ssr-csr': path.join(__dirname, 'ssr-demo', 'dist-csr'),
  '/ssr-ssr': path.join(__dirname, 'ssr-demo', 'dist-ssr'),
  '/islands': path.join(__dirname, '..', 'examples', 'islands', 'dist'),
};

const MIME = {
  '.html': 'text/html; charset=utf-8',
  '.js': 'text/javascript; charset=utf-8',
  '.mjs': 'text/javascript; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.wasm': 'application/wasm',
  '.json': 'application/json',
  '.svg': 'image/svg+xml',
  '.map': 'application/json',
};

function pickRoot(urlPath) {
  for (const [prefix, dir] of Object.entries(ROOTS)) {
    if (urlPath === prefix || urlPath.startsWith(prefix + '/')) {
      return { prefix, dir };
    }
  }
  return null;
}

export function startServer(port = 0) {
  const server = http.createServer(async (req, res) => {
    try {
      const url = new URL(req.url, 'http://localhost');
      const match = pickRoot(url.pathname);
      if (!match) {
        res.writeHead(404).end('no app at ' + url.pathname);
        return;
      }
      let rel = url.pathname.slice(match.prefix.length);
      if (rel === '' || rel === '/') rel = '/index.html';
      const filePath = path.join(match.dir, decodeURIComponent(rel));
      if (!filePath.startsWith(match.dir)) {
        res.writeHead(403).end('forbidden');
        return;
      }

      let st;
      try {
        st = statSync(filePath);
      } catch {
        res.writeHead(404).end('not found: ' + rel);
        return;
      }
      if (st.isDirectory()) {
        res.writeHead(404).end('is a directory');
        return;
      }

      const ext = path.extname(filePath);
      const type = MIME[ext] || 'application/octet-stream';
      const etag = `"${st.size}-${Math.floor(st.mtimeMs)}"`;

      if (req.headers['if-none-match'] === etag) {
        res.writeHead(304, { ETag: etag, 'Cache-Control': 'public, max-age=300' }).end();
        return;
      }

      const acceptsGzip = (req.headers['accept-encoding'] || '').includes('gzip');
      const headers = {
        'Content-Type': type,
        'Cache-Control': 'public, max-age=300',
        ETag: etag,
        // cross-origin isolation so performance.measureUserAgentSpecificMemory()
        // is available in the runner. All resources here are same-origin, so
        // require-corp is safe.
        'Cross-Origin-Opener-Policy': 'same-origin',
        'Cross-Origin-Embedder-Policy': 'require-corp',
      };

      if (acceptsGzip) {
        const buf = await fs.readFile(filePath);
        const gz = zlib.gzipSync(buf, { level: 9 });
        headers['Content-Encoding'] = 'gzip';
        headers['Content-Length'] = gz.length;
        res.writeHead(200, headers);
        res.end(gz);
      } else {
        headers['Content-Length'] = st.size;
        res.writeHead(200, headers);
        createReadStream(filePath).pipe(res);
      }
    } catch (err) {
      res.writeHead(500).end(String(err));
    }
  });

  return new Promise((resolve) => {
    server.listen(port, () => resolve({ server, port: server.address().port }));
  });
}

// Run standalone: `node server.mjs [port]`
if (import.meta.url === `file://${process.argv[1]}`) {
  const port = parseInt(process.argv[2] || '8090', 10);
  startServer(port).then(({ port }) => {
    console.log(`bench server on http://localhost:${port}/  (mounts: ${Object.keys(ROOTS).join(', ')})`);
  });
}
