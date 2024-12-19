const { createProxyMiddleware } = require('http-proxy-middleware');

module.exports = function(app) {
  app.use(
    '/api',
    createProxyMiddleware({
      target: 'http://hactiv-web-backend:8080',
      changeOrigin: true,
    })
  );
};
