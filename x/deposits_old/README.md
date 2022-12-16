# Deposits
## Overview
The deposits module contains functionality for tracking blockchain events, interacting with smart contracts, and managing wallet balances.  The module listens to events coming from the blockchain, feeds these events to Kafka which are consumed by the module's processor, and can interact with the blockchain using the chain client.

## Chain Client
The chain client is a wrapper around different blockchain clients.  The intent of the wrapper is to make the deposit module's interactions with any blockchain client be genereic as possible to make adding chains / altering deposit contracts easy.  Currently, the chain client is specifically responsible for:
- Managing the event subscription for new blocks
- Getting the latest block number
- Providing a contract interface

When new events are received from the chain client, they are sent via the Request Service (Kafka producer) to the processor.  All sent events are asynchronous calls, and the chain client does not wait for a response from the processor.  See info at the bottom of this page on how different events are serialized and sent to the processor.

### Contract Interface
The contract interface is a the main interface for interacting with specific blockchain events.  The interface is a wrapper around the generated contract bindings.  For EVM chains, these bindings are generated using the [abigen](https://geth.ethereum.org/docs/install-and-build/installing-geth) tool. The contract interface is responsible for:
- Getting deposit events.  It filters events by a range of blocks, and the intended deposit recipient (the address of a node).
- Getting withdrawal confirmations.  Once a node returns funds to a user on the blockchain, it will wait to hear for a confirmation from the smart contract.  Similar to deposits, the contract interface filters events by a range of bhttps://flaviocopes.com/golang-data-structure-binary-search-tree/locks, and the intended recipient (the node).
- Returning funds.  The ReturnFunds function initiates the withdrawal process, returning blockchain tokens held in escrow back to the user.

## Event Feed
The event feed manages the subscriptions to the chain client.  Currently, this is only listening for new blocks and ensures that they are not a soft-fork / will not become soft-forked, but this listener could be used for theoretically any event.

The event feed will listen to new blocks as they are mined onto the blockchain, and keep them in a queue until there is probabalistically no soft-forks.  The event feed then emits these block headers, from which the deposit module will get all relevant events from that block height.  The event feed also manages reconnection logic, and blocks that are submitting in faulty order (e.g., if we received block number 49->50->53, then it can recover blocks 51 and 52, and emit them in order).

## Structures
Structures contains 3 different data structures used by the deposits module.  I separated these out in case they are useful for other modules, but they are not currently used by any other module.

The first is a very basic thread-safe FIFO queue.  The queue is used to manage blocks that have been mined but have not reached confirmation yet.

The second is a thread-safe binary search tree.  The BST is used to store pending withdrawals, sorted by their expiraton height.  It does not have all of the methods that a BST would usually have (e.g. max, traversals, etc), since I only implemented the methods that I needed.  More can be implemented as needed.  Most of it was taken from [here](https://flaviocopes.com/golang-data-structure-binary-search-tree/).  I could not find a license.

The third structure is a structure for managing withdrawals, which is simply a binary search tree that also contains a hash map with a nonce pointing to the node in the search tree.  This allows us to quickly find a node in the tree, either by its expiration time (block height), or by its nonce (in the case it is confirmed before expiration).  This structure is **NOT** thread-safe, and should only be used by a single thread.  It can be made thread-safe fairly easily.

## Config
The deposit-config.yaml file is used to configure the deposit module.  It contains the following fields:

- required-confirmations: The number of blocks that must be mined after a deposit is made before it is considered safe from soft forks.
- block-timeout: The number of seconds where the deposit module will reconnect to the chain client if it has not received a new block.
- withdrawal-expiration: The number of blocks that a withdrawal can be in pending for.
- chain: The client chain identifier.  See below for more details.
- provider-endpoint: The endpoint for an RPC provider.  This is used to connect to the chain client.
- contract-address: The address of the deposit contract.
- keys:
    - key-path: path to a PEM file containing a private key.
- sync:
    - chunk-size: The number of blocks to get events for at a time.

## Supported Chains
Currently Kwil can technically suppot EVM chain, but is only set up to handle a few.  Below is a table of chains, the Kwil chain identifier, and the chain ID used by the chain.

| Chain | Kwil Chain Identifier | Chain ID |
|-------|-----------------------|----------|
| Ethereum | eth-mainnet | 1 |
| Goerli | eth-goerli | 5 |

## Event Serialization
Events sent from the deposit module to Kafka are serialized according to a predeterimined schema.  This enables the consumer to identify the type of event, deserialize into the correct struct, and handle the event accordingly.  Serialized events start with a magic byte set as 0.  This is included in case we need to add more event types later.  The second byte identifies the event type:

Magic Byte | Event Type | Data

Below is a table of the different events, their byte indicator, and the struct they are deserialized into.

| Event Type | Byte | Struct |
|------------|------|--------|
| Deposit | 0 | DepositEvent |
| Withdrawal Request | 1 | WithdrawalEvent |
| Withdrawal Confirmation | 2 | WithdrawalConfirmationEvent |
| End of Block | 3 | EndOfBlockEvent |
| Spend | 4 | SpendEvent |

## Store
**This is not currently being used, and has been replaced with Kafka + the processor**
The store is responsible for persistently storing user's deposit, spent, and pending withdrawal values.