We need to prepare migration script first and using upgrade cli to manipulate it.

## 1. **Get upgrade module-versions**

   Clone our forked repository:

   ```bash
   aizeld query upgrade module-versions
   ```
  check the version with module name is "evm"

## 2. **Write a New EVM Migration (v9)**

   Create a new migration package (for example, x/evm/migrations/v9) that contains a migration function which transforms the stored state from the old version (v8) to your new version (v9) with your added EIP‑5656 changes (MCOPY opcode support).

## 3. **Write a New Upgrade**

   Mae the upgrade file in app/upgrades/evm-v9/upgrades.go

## 4. **Register the Upgrade Handle**
In your application’s upgrade setup function (often called setupUpgradeHandlers() in your app file), register the new handler. For example, if your application holds the evm store key in a field (say, app.keys[evmtypes.StoreKey]), you would do something like:
```golang
   func (app *Evmos) setupUpgradeHandlers() {
    // v20 upgrade handler (existing)
    app.UpgradeKeeper.SetUpgradeHandler(
        v20.UpgradeName,
        v20.CreateUpgradeHandler(app.mm, app.configurator, app.AccountKeeper, app.EvmKeeper, app.appCodec),
    )

    // EVM v9 upgrade handler
    app.UpgradeKeeper.SetUpgradeHandler(
        v9.UpgradeName,
        v9.CreateUpgradeHandler(app.mm, app.configurator, app.AccountKeeper, app.EvmKeeper, app.appCodec, app.keys[evmtypes.StoreKey]),
    )

    // (Other upgrade handling code...)
}
```

## 5. **Schedule the Upgrade via governance**

Below is an example of how you can schedule a software upgrade via governance on a Cosmos‑SDK–based chain (like your Evmos‑based chain). In Cosmos‑SDK, you do this by submitting a **SoftwareUpgradeProposal** via the governance module. When the proposal passes, the upgrade plan (with your upgrade name and target block height) is set, and when that block is reached the upgrade handler is triggered.

### 1. Create a Proposal JSON File

Create a JSON file (for example, `evm-v9-upgrade-proposal.json`) with the following content. Adjust the parameters (title, description, height, info URL, deposit amounts) to your requirements:

```json
{
  "title": "Upgrade EVM Module to v9 (EIP-5656)",
  "description": "This proposal upgrades the EVM module to v9 to enable EIP-5656 (the MCOPY opcode). This upgrade includes the necessary state migrations.",
  "plan": {
    "name": "evm-v9",
    "height": "1234567",
    "info": "https://github.com/YourRepo/YourUpgradeInfo"
  }
}
```

*Notes:*
- **name:** Must match the upgrade name you registered in your upgrade handler (e.g. `"evm-v9"`).
- **height:** Specify the block height at which the upgrade should take effect.
- **info:** A URL pointing to additional upgrade details (this is optional but recommended).

### 2. Submit the Proposal via CLI

Use the governance transaction command to submit your proposal. For example, using your chain’s CLI (here assumed to be `aizeld`):

```bash
aizeld tx gov submit-proposal software-upgrade evm-v9-upgrade-proposal.json \
  --from <your-key> \
  --deposit 1000000stake \
  --chain-id <your-chain-id> \
  --fees 200000stake \
  --yes
```

Replace:
- `<your-key>` with your key or account name.
- `<your-chain-id>` with your chain ID.
- The deposit and fee amounts with values appropriate for your chain’s parameters.

### 3. Vote on the Proposal

Once the proposal is submitted, make sure that enough votes are cast for it to pass. You can check the proposal status with:

```bash
aizeld query gov proposal <proposal-id>
```

Then, vote on the proposal if necessary:

```bash
aizeld tx gov vote <proposal-id> yes --from <your-key> --chain-id <your-chain-id> --fees 200000stake --yes
```

### 4. Upgrade Execution

When the chain reaches the specified block height, the upgrade handler will be triggered automatically. Your new binary (compiled with the evm‑v9 changes) must be running at that point so that the migration code (for example, enabling EIP‑5656) is executed.