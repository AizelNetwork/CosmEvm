## 1. Prerequisites

We need to prepare migration script first and using upgrade cli to manipulate it.

1. **Get upgrade module-versions**

   Clone our forked repository:

   ```bash
   aizeld query upgrade module-versions
   ```

2. **Build and Install the Binary**

   Use the provided Makefile to build and install the binary:

   ```bash
   make install
   ```

   > **Tip:** If you run into issues related to linker flags, run the `unset LDFLAGS` and `unset CFLAGS` commands before `make install`.

---

## 2. Configure Node1

### 2.1. Generate Bech32 Format Addresses from EVM Cold Wallets

Your chain uses Bech32‑formatted addresses. For any external (cold wallet) EVM address you want to fund, convert the EVM hex address into Bech32 with the following command:

```bash
aizeld debug addr [evm-address]
```

For example:

```bash
aizeld debug addr 0xaaafB3972B05630fCceE866eC69CdADd9baC2771
```

This outputs something like:

```
Address bytes: [170 175 179 151 43 5 99 15 204 238 134 110 198 156 218 221 155 172 39 113]
Bech32 Acc: aizel142hm89etq43sln8wsehvd8x6mkd6cfm35279kg
Bech32 Val: aizelvaloper142hm89etq43sln8wsehvd8x6mkd6cfm3k9mj49
```

Use the **Bech32 Acc** address (starting with `aizel1`) for allocating genesis coins to your external wallets. Update the addresses for `USER1`, `USER2`, `USER3`, and `USER4` in your production script (`prod_node1.sh`) accordingly.

### 2.2. Customize Validator Mnemonics

In your production initialization script (`prod_node1.sh`), you will see variables such as `VAL1_MNEMONIC` and `VAL2_MNEMONIC`. **These are sample mnemonics provided for testing purposes only.**  
For production, **customize these values** with your own secure mnemonic phrases:

```bash
VAL1_MNEMONIC="your secure mnemonic for validator1 goes here"
VAL2_MNEMONIC="your secure mnemonic for validator2 goes here"
```

Make sure to keep these mnemonics safe and never share them publicly.

### 2.3. Initialize Node1

Run your production initialization script for node1:

```bash
./prod_node1.sh
```

Your `prod_node1.sh` should perform tasks such as:
- Setting client configuration (chain-id, keyring, etc.)
- Importing validator keys using your custom mnemonics
- Initializing the node (`aizeld init`)
- Adjusting denominations in the genesis file
- Allocating genesis accounts (using your Bech32 addresses)
- Generating a gentx for validator1

After running the script, verify the genesis file:

```bash
aizeld validate-genesis --home $AIZELHOME/node1
```

---

## 3. Configure Node2

Your production script for node2 (`prod_node2.sh`) creates a second node by copying the node1 folder and then modifying settings. It also signs a gentx for validator2.

- Deleting any existing node2 folder  
- Copying node1’s folder to node2  
- Changing file ownership (so your user can modify them)  
- Updating ports, node name, and persistent peers  
- Signing validator2’s gentx  
- Copying the gentx files from node2 back into node1  

```bash
./prod_node2.sh
```

> **Reminder:** Update your validator mnemonics in your production scripts (e.g. in `prod_node1.sh` and any related configuration files) with your own secure values.

---

## 4. Collect Genesis Transactions

After node1 and node2 have signed their gentxs, run the genesis collection script to aggregate all gentx files and update the genesis file.

Create a script called `collect-gentxs.sh` (or update your existing one) with the following content:

```bash
./collect-gentxs.sh
```

---

## 5. Start the Nodes

Create a start script (`start-nodes.sh`) that starts both nodes. For example:

```bash
./start-nodes.sh
```

---

## 6. Test Your Chain

After the nodes have started, test your blockchain by retrieving the current block number via the JSON‑RPC endpoint:

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://localhost:18545
```

You should receive a JSON response with the block number (in hexadecimal).

---

## FAQ

**Q:** *I encountered an issue when running `make install` regarding OpenBLAS linker flags.*  
**A:** Unset the conflicting environment variables before building:

```bash
unset LDFLAGS
unset CFLAGS
```
---
