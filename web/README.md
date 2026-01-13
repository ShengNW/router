# React Template

## 快速启动（配合后端）

```bash
# 后端（在仓库根目录）
AUTO_REGISTER_ENABLED=true WALLET_JWT_SECRET=dev_secret go run ./cmd/router --log-dir ./logs

# 前端（本目录）
npm install
npm start
```

默认会代理到 http://localhost:3011。

## Basic Usages

```shell
# Runs the app in the development mode
npm start

# Builds the app for production to the `build` folder
npm run build
```

If you want to change the default server, please set `REACT_APP_SERVER` environment variables before build,
for example: `REACT_APP_SERVER=http://your.domain.com`.

Before you start editing, make sure your `Actions on Save` options have `Optimize imports` & `Run Prettier` enabled.

## Reference

1. https://github.com/OIerDb-ng/OIerDb
2. https://github.com/cornflourblue/react-hooks-redux-registration-login-example
