/** @type {import('http-proxy-middleware').Options} */
const bypass = (req) => {
  // Let browser navigation (document) requests fall through to Angular's
  // historyApiFallback so the SPA is served instead of the API JSON.
  if (req.headers['accept']?.includes('text/html')) return '/index.html';
};

module.exports = {
  '/products': { target: 'http://localhost:8080', secure: false, changeOrigin: true, bypass },
  '/customers': { target: 'http://localhost:8080', secure: false, changeOrigin: true, bypass },
  '/orders': { target: 'http://localhost:8080', secure: false, changeOrigin: true, bypass },
  '/health': { target: 'http://localhost:8080', secure: false, changeOrigin: true, bypass },
};
