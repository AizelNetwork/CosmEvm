### 1. Initial node1 :

#### 1.1. Generate bench32 format address from evm cold wallet address.

```bash
aizeld debug addr [evm-address]
```

apply the above generated bench32 address for the USER1, USER2, USER3, USER4 in file prod_node1.sh


#### 1.2. Initial node1 by running script :

```bash
./prod_node1.sh
```

### 2. Do with node2 

### 2.1 Initial node1 by running script :

```bash
./prod_node2.sh
```

### 3. Back with node2 
#### 3.1 Collect genesis tx 

```bash
./collect-gentxs.sh
```

#### 3.3 Start the nodes

```bash
./start-nodes.sh
```

### 3.4 Get block number to test the chain

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://localhost:18545
```

## FAQ
###If you meet this issue when run "make install" : flag provided but not defined: -L/opt/homebrew/opt/openblas/lib, Just need to do :

```bash
unset LDFLAGS
unset CFLAGS
```
