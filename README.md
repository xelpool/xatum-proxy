# XATUM-PROXY v0.1.1
An open-source, high-performance Xatum mining proxy.
Designed to split the work between multiple miners and to offer a bridge between Xatum and Getwork.

## Usage
1. Download XATUM-PROXY and extract it, then run it
2. Start your miner of choice and point it to the proxy's `IP:PORT`.

Example XATUM-PROXY mining with xelis_miner:
```
./xelis_miner -m YOUR_WALLET_ADDRESS --daemon-address 127.0.0.1:5210
```
Example XATUM-PROXY mining with 3DP-The-AllFather/xelis-gpu-miner
```
./xelis-taxminer --wallet YOUR_WALLET_ADDRESS --host 127.0.0.1:5210 --boost
```

## Command-line flags
- `--wallet <WALLET ADDRESS>`: Starts XATUM-PROXY with the given wallet address
- `--debug`: Starts in debug mode