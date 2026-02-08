# Running the Monorepo

This repo contains three components:

* `ember-contracts` (Sui Move contracts)
* `ember-frontend` (React app)
* `ember-lender` (Go service)

---

## 1. Prerequisites

Install:

* Node.js (>= 18)
* Go (>= 1.21)
* Sui CLI
* pnpm / npm / yarn

Configure Sui:

```bash
sui client new-address ed25519
sui client switch --address <YOUR_ADDRESS>
```

---

## 2. Run Contracts (`ember-contracts`)

```bash
cd ember-contracts

sui move build
sui client publish --gas-budget 100000000
```

Save:

* Package ID
* Deployed object IDs

---

## 3. Run Backend Node (`ember-lender`)

```bash
cd ember-lender

go mod tidy
go run main.go
```

Set env variables:

```bash
export SUI_RPC_URL=https://fullnode.devnet.sui.io
export SUI_PRIVATE_KEY=<YOUR_PRIVATE_KEY>
export PACKAGE_ID=<DEPLOYED_PACKAGE_ID>
```

---

## 4. Run Frontend (`ember-frontend`)

```bash
cd ember-frontend

npm install
npm run dev
```

Set environment variables:

```bash
VITE_SUI_RPC_URL=https://fullnode.devnet.sui.io
VITE_PACKAGE_ID=<DEPLOYED_PACKAGE_ID>
VITE_MANAGER_OBJECT_ID=<OBJECT_ID>
```

---

## 5. Run All (3 terminals)

Terminal 1:

```bash
cd ember-contracts && sui client publish
```

Terminal 2:

```bash
cd ember-lender && go run main.go
```

Terminal 3:

```bash
cd ember-frontend && npm run dev
```

---

## 6. Networks

Default: `testnet`

To switch:

```bash
sui client switch --env devnet | testnet | mainnet
```
