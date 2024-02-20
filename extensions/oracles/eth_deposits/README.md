# Ethereum-based Token Deposit Oracle

EVM-based Token Deposit oracle is used to credit accounts on a Kwil network based on the events emitted from EVM chains. This oracle actively listens for the `Credit(address,uint256)` event signatures and credits the accounts on the Kwil network upon attestation by a super-majority (two-thirds) of Validators confirming the occurrence of this event.

This oracle facilitates a variety of use-cases, such as below to acquire Kwil gas:

- Escrow ERC20 Tokens:
  - By depositing ERC20 tokens into escrow, users can acquire Kwil gas, with the tokens being securely held in the smart contract.
- Burn ERC20 Tokens:
  - Users can burn ERC20 tokens to acquire Kwil gas.
- NFT Minting Incentives:
  - Minters of NFTs can be rewarded with Kwil gas as a part of the minting process.

## Sample Smart Contract

Below is a sample smart contract defining a `sampleMethod` emitting `Credit` event. This smart contract serves as a basic example and does not include actual interactions with ERC20 tokens or NFTs, focusing instead on the structure and the critical `Credit` event emission. Deploy the contract onto the EVM chain once it's ready.

```solidity
// SPDX-License-Identifier: GPL-3.0
pragma solidity >=0.8.0;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract Escrow {
  
    event Credit(address _from, uint256 _amount);

    IERC20 public escrowToken;

    constructor(address _escrowToken) {
        escrowToken = IERC20(_escrowToken);
    }

    /**
        Sample method that emits the Credit event, which credits the account 
        on the Kwil network with the amt specified.
     */
    function sampleMethod(uint256 amt) public {
        /* Your code */

        emit Credit(msg.sender, amt);
        
        /* Your code */
    }
}
```

## Enable EVM-based Token Deposit Oracle On Kwil Node

The below configuration is required in `config.toml` to enable `eth_deposits` oracle on a Kwil node.

```yaml
[app.oracles.eth_deposits]

"rpc_provider": "wss://sepolia.gateway.tenderly.co"
"contract_address": "0x94e6a0aa8518b2be7abaf9e76bfbb48cab1545ad"
"starting_height": "83100"
"required_confirmations": "12"
"reconnection_interval": "30"
"max_retries": "10"
"block_sync_chunk_size": "1000000"
```

1. `rpc_provider`: This variable specifies the WebSocket URL of the EVM node provider. This is a required field and would likely be an Infura or Alchemy URL.
2. `contract_address`: This variable specifies the address of the deployed smart contract for the oracle to listen to the `Credit` events at. This is a required field.
3. `starting_height`: This variable specifies the Ethereum block height at which the oracle starts to listen for the `Credit` events. Any events emitted before this block height are ignored. The default value is `0`.
4. `required_confirmations`: This variable specifies the number of ethereum blocks that must be mined before the oracle creates a deposit event in Kwil. This is to protect against Ethereum network reorgs or soft forks. This is an optional field and defaults to `12` if not configured.
5. `reconnection_interval`: This variable specifies the amount of time in seconds that the oracle will wait before reconnecting to the Ethereum RPC endpoint if it is disconnected. Long-running RPC subscriptions are prone to being reset by the Ethereum RPC endpoint, so this will allow the oracle to reconnect. This is an optional field and defaults to `60`s if not configured.
6. `max_retries`: This variable specifies the number of times the oracle will attempt to reconnect to the Ethereum RPC endpoint before giving up. This is an optional field and defaults to `10` if not configured.
7. `block_sync_chunk_size`: This variable specifies the number of Ethereum blocks requested by oracle from the Ethereum RPC endpoint at a time while catching up to the network. This is an optional field and defaults to `1000000` if not configured.
