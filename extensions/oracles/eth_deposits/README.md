# Ethereum-based Token Deposit Oracle

EVM-based Token Deposit oracle is used to credit accounts on a Kwil network based on the events emitted from EVM chains. This oracle actively listens for the `Credit(address,uint256)` event signatures and credits the accounts on the Kwil network upon attestation by a super-majority (two-thirds) of Validators confirming the occurrence of this event.

This oracle facilitates a variety of use-cases, such as below to acquire Kwil gas:

- Escrow ERC20 Tokens:
  - Through smart contracts that emit `Credit` events, users can acquire network gas using ERC20 tokens, which are subsequently stored within the smart contract.
- Burn ERC20 Tokens:
  - Users can burn ERC20 tokens to acquire Kwil gas.
- NFT Minting Incentives:
  - Minters of NFTs can be rewarded with Kwil gas as a part of the minting process.

## Sample Smart Contract

Below is a sample Escrow Contract defining a `sampleMethod` emitting `Credit` event. This smart contract serves as a basic example and does not include actual interactions with ERC20 tokens or NFTs, focusing instead on the structure and the critical Credit event emission. Before deploying this contract onto an EVM chain, ensure to integrate it with actual ERC20 token handling or NFT minting logic, as well as thorough testing and security audits to prevent vulnerabilities. Deploy the Escrow Contract onto the EVM chain once it's ready.

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

The below configuration is required in `config.toml` to enable `eth_deposit_oracle` oracle on a Kwil node.

```yaml
[app.oracles.eth_deposit_oracle]

"endpoint": "wss://sepolia.gateway.tenderly.co"
"chain_id": "11155111"
"escrow_address": "0x94e6a0aa8518b2be7abaf9e76bfbb48cab1545ad"
"starting_height": "83100"
"required_confirmations": "12"
```

1. `endpoint`: This variable specifies the WebSocket URL of the EVM node provider.
2. `chain_id`: This variable specifies the ChainID of the Ethereum node provider.
3. `escrow_address`: This variable specifies the deployed smart contract address for the oracle to listen to the `Credit` events at.
4. `starting_height`: This variable specifies the starting block height at which the oracle begins to listen for `Credit` events. The default value is `0`. Any events emitted before this block height are ignored.
5. `required_confirmations`: This variable specifies the number of block confirmations required for a block to be finalized. The default value is `12`.
