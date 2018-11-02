#!/usr/bin/env bash

npm install webpack babel-core babel-plugin-transform-react-jsx babel-preset-env --save-dev
npx babel --plugins transform-react-jsx src/dashboard.js --out-file src/index.js
npx webpack
