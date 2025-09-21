// http://eslint.org/docs/user-guide/configuring

const path = require('path');

const resolve = dir => path.resolve(__dirname, dir);

module.exports = {
  root: true,
  parserOptions: {
    project: ['./tsconfig.json'],
  },
  extends: ['kubesphere'],
  settings: {
    'import/resolver': {
      webpack: {
        config: resolve('node_modules/@ks-console/bootstrap/webpack/webpack.base.conf.js'),
      },
    },
  },
};
