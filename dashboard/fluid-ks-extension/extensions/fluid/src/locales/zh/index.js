// Use require.context to require reducers automatically
// Ref: https://webpack.github.io/docs/context.html
const context = require.context('./', false, /\.json$/);
const keys = context.keys().filter(item => item !== './index.ts');

// let models: Record<string, string> = {};
let models = {};
for (let i = 0; i < keys.length; i += 1) {
  if (keys[i].startsWith('.')) {
    models = { ...models, ...context(keys[i]) };
  }
}

export default models;
