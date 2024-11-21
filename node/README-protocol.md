# Protocol Draft Notes

## libp2p

The main concepts in `libp2p` are: peers, streams, and protocols.

### Multiplexed Peer Connection

A connection to a peer is a persistent multiplexed connection.

- connection is persistent (connect/close)
- connection is multiplexed (concurrent communication protocols isolated with "streams")
- connection can use any transport or security (plain, tls, noise)
- peer ID is like address, derived from public key, authenticates peer

### Streams and Protocols

Within a multiplexed connection, there are any number of "streams" that isolate a communication "protocol".

- streams may be short lived or long lived, and are inexpensive to create
- a stream has a certain protocol: established message sequence that is application-defined and dictated by application code.
- a protocol is specified when opening the stream
- the actual protocol is transparent to libp2p, defined by application
- the expected sequence of messages, their encodings, and logic is what defines a protocol
- peer defines a "stream handler" to allow a peer to initiate a stream with a particular protocol
- peer creates a new stream instance for with a specific (1) peer, and (2) protocol
- either side of the new stream may initiate messaging, but it is normally the stream creator, not the stream handler
- the stream is multiplexed on the p2p connection with other steams, thus isolated and concurrent
- stream handlers and initiators define all timeouts, limits, etc. as the `Stream` is essentially an `io.ReadWriteCloser`, with other methods to get context like the remote peer

#### Example 1: `ProtocolIDTx`

This protocol is a simple request/response for getting the contents of a transaction.

```go
ProtocolIDTx       protocol.ID = "/kwil/tx/1.0.0"
```

The application currently defines the protocol as follows:

1. Stream creator sends a message prefixed with

  ```go
  getTxMsgPrefix  = "gettx:"
  ```

  If the stream creator wants a transaction ID `"1c577d897bb6cef3cffb8a6e323289eec5c85024dacd188d485fbf3bb003bb76"`, it `Write`s to the stream the message `"gettx:1c577d897bb6cef3cffb8a6e323289eec5c85024dacd188d485fbf3bb003bb76"`.

2. The stream handler waits for the message i.e. a blocking `Read`.

3. The stream handler parses the message (detect the prefix, and decode the txid), fetches the transaction.

4. If the transaction is not known or available, `Write` a `noData` message (just a `"0"` or something defined under this protocol).
   
5. If the transaction is known and available, `Write` the bytes of the transaction.

6. The stream creator waits for the response i.e. a blocking `Read`.

7. Both sides close the stream.

#### Example 2: `ProtocolIDTxAnn`

This protocol is of a dialog, although still relatively short-lived.

```go
ProtocolIDTxAnn    protocol.ID = "/kwil/txann/1.0.0"

annTxMsgPrefix  = "txann:"
```

The function `advertiseTxToPeer` creates the new stream to advertise the transaction.

Stream handler function `txAnnStreamHandler` handles the solicitation.

1. The announcer peer creates a stream to its peers.

2. It writes the message `"txann:1c577d8...` (full txid)

3. The stream handler peer reads and parses the message, and checks its inventory to see if it has or needs the transaction.

4. If the peer has the transaction already, it closes the stream. Fin.

5. If the peer needs the transaction content, it sends back a `"get"` message (on the same live stream).

6. The advertising peer recognizes the `"get"` response and sends the transaction contents.

7. The receiving peer receives and stores the transaction contents, and closes the stream.

The protocol ends there. At this point in the application, the stream handler would initiate async re-announce to other peers.

#### Other streams

There are also similar protocols for blocks, and a `ProtocolIDDiscover` for peer discovery with streams.

The peer discovery protocol is used by application methods that periodically request peer lists from its connected peers, which are added to its own peer store.  We will have to build logic around the peer store e.g. max nodes, types of nodes to connect to (only validators from leader, etc), etc.

The stream handler for `ProtocolIDDiscover` would just close the stream if the node had pex disabled. If enabled, the stream handler just sends it peer list on request. Here we can also build application logic, such as do not send certain peers depending on the identity of the requester.

## Kwil Protocols for Transactions, Blocks, and Consensus

### Overall Kwil Requirements

The following requirements are for the overall Kwil blockchain network design, and will dictate the specific P2P protocol requirements in the next section.

The network SHOULD ensure that the block producer (the leader) has all known transactions.

Blocks produced by the leader MUST propagate to the entire network. Validators SHOULD obtain the blocks quickly / with priority.

Consensus ACKs/NACKs from the validators SHOULD be returned to the leader quickly.

Leader COMMIT/ROLLBACK/ABRT should be announced and reach validators quickly.

Consensus messages MAY relay promiscuously (even through non-validators) to propagate quickly, but only validators process the messages.

All consensus and block messages MUST be signed and authenticated appropriately (blocks from leader, ACK/NACK from validator, COMMIT/ROLLBACK/ABRT from leader, etc.)

Transactions SHOULD be signed and/or validated prior to (re)announce, ingest into mempool, or included in block by leader.

To facilitate initial block download (IBD) to sync a node (outside of consensus process), peers MUST be able to:

- request blocks, sequentially or in batch
- handle block requests from all nodes, prioritizing validators (or recovering leader)
- sanity check retrieved block prior to executing it (height, prev block hash, leader signature, other header elements)

To facilitate robust consensus in restart/reconnect scenarios, leader and validator:

- validators SHOULD re-send/broadcast ACK/NACK (containing their computed app hash for the block), conditionally if that block is latest
- validators SHOULD be able to solicit the leader's consensus status for a block (waiting for ACKs, or block already committed/aborted) -- maybe not?
- leader SHOULD re-send COMMIT/ROLLBACK/ABRT periodically if insufficient responses received by a deadline
- leader SHOULD respond to any out-of-sequence messages as appropriate for the validator to resolve e.g. ACK for a previous or unknown block proposal can be met with a RSLV message with current block and consensus stage
- (?) leader MAY propose a new block when insufficient responses (given timeout)

### Kwil P2P Requirements

- all nodes must announce new transactions created locally (e.g. received from RPC or authored by the node)
- nodes should retrieve unknown transactions that were announced to it
- nodes should re-announce the transactions that it has retrieved and which pass basic validity checks
- nodes should periodically re-announce unconfirmed transactions
- non-leader nodes should perform peer exchange to mesh for efficient content distribution, validators prioritizing other validators
- leader must broadcast new block IDs, and serve on request
- all nodes must be able to request specific blocks, by height or hash (maybe other identifiers)
- nodes may operate in "blocks only" mode, such as when syncing or when functioning as an archival node with no mempool
- leader should only accept connections from (peer with) validators, so as to prevent spamming on the leader

Other considerations arise from semantics of mempool and block confirmation. For instances, remove transactions from mempool when they are mined, and re-check unconfirmed transactions after confirming a blocks, etc.  However, this section pertains to expected p2p behavior to facilitate function of Kwil in the leader-based block production design.

### Stream Protocol Specs

The above requirements necessitate a number of stream protocols, which are detailed here.

The protocol stream handlers and initiator methods will interact with several other higher level systems including: mempool, block store, block index, transaction index, consensus engine, etc.

Non-blocking: All stream handlers and initiators MUST NOT block other node functions, and MUST be able to run concurrently. Other systems that interact with the P2P layer SHOULD utilize atomics, queues, goroutines, and other internal mechanisms for handling communications *asynchronously*.

Out-of-sequence or otherwise invalid/unexpected messages received on incoming (remotely initiated) streams or from any higher level gossip protocols should be either ignored or initiate resolution of consensus state. For instance, if a validator may reset to the "waiting for proposal" consensus state or use a resolution protocol to rejoin the current round. These scenarios may arise if a validator looses connectivity or a temporary network partition interrupts leader messages, or if the validator simply starts up mid-round.

The following stream protocols are needed, where there is a stream initiator (does `NewStream`, outgoing to peer) and a stream handler (had `SetStreamHandler` pointing to a handler function, for incoming from a peer).

#### Data retrieval protocols

- `ProtocolIDBlock` - Block get
  - summary: request block content by hash or height, perhaps separate protocols
  - initiator: any peer type
  - handler: any peer type
  - context note: this protocol is important to sync, and requesting block content from peers other than the advertising peer (see below)

- `ProtocolIDTx` - Tx get
  - summary: request tx content by txid
  - initiator: any peer type
  - handler: any peer type
  - context note: most txns will be retrieved on the announce protocol (see below), but this can be used to request the content from a different peer

#### Content announcement protocols

- `ProtocolIDTxAnn` - Tx announce
  - summary: announce txid available, serve content if requested
  - initiator: any peer type, but leader would only announce their own (not re-announce others)
  - handler: any peer type (all must do this for txns to reach leader)
  - note: this protocol involves re-announce from non-leader peers, effectively making this a custom gossip protocol, to ensure leader is not burdened with re-announce

- `ProtocolIDBlockAnn` - (committed) block announce
  - summary: announce a committed/finalized block, which must have leader signature
  - initiator: any peer type, but originates from leader
  - handler: any peer type except leader, where sentry is likely to request the block content, and validator would already have executed it from the preceding proposal (see below)
  - context notes: for a sentry, this is just a new block to get and execute; for a validator, this is the final "commit" step in the consensus sequence. This protocol also involves re-announce to propagate the block. We may also use gossipsub so the message signature from the initiator (leader) is retained, but the announce would no longer imply direct content availability from the peer who relayed it, necessitating more complex content discovery (dht, asking peers, retry logic, etc.).

#### Leader-validator protocols

- `ProtocolIDBlockPropose` - proposed block advertisement
  - summary: announce a new block proposal, leader signed, which validators should fetch (on the same stream) and execute (async, after closing the stream)
  - initiator: leader or validator (if relaying to other validators), but content originates from leader. or could require direct leader->validator stream
  - handlers: validators only, sentry must ignore and validators should not send to sentry

- `ProtocolIDACKProposal` - post-exec validator result
  - summary: validators result of executing a proposed block, ACK(+appHash)/NACK, is provided to the leader
  - initiator: validator
  - handler: leader
  - may also be gossip protocol instead of p2p stream since the message is small with only a hash and a signature

NOTE: If block execution were always very fast (<1 sec), the two above could be a single protocol using one longer-lived stream. Reconnects, however brief, break the stream.

maybe:

- `ProtocolIDProposedBlock` - if validator wants to request a block proposal outside of the advert stream (`ProtocolIDBlockPropose`), such as if they lost the connection before pulling the entire proposed block contents in that stream. Probably just miss the round or wait for re-announce from leader instead of this. But this does seem very simply to implement like the other content retrieval protocols described above.
- `ProtocolIDACKCommit` - response from validator to leader after commit, needed? There may not need to be any response. Also, since doing the Commit is fast if they already executed the proposal, any response would just be in the same stream (`ProtocolIDBlockAnn`) unlike the proposal ACK/NACK that could take time for a validator to produce.

## Appendix A: Block Production Stages

### Leader's perspective

There are fundamentally just two stages:

1. `PROPOSED` -- new block proposed by leader, collecting validator ACKs/NACKs, concurrently executing the block.

   Once a threshold of validator ACK appHashes match the locally computed appHash (and prepared transaction is ready to commit), commit the changes.

2. `COMMITTED` -- announce the committed block to validators, now idle or assembling the next block

There are multiple concurrent processes and statuses within `PROPOSED` (announcing the proposal, locally executing, and receiving validator results), but in terms of block production, this can be thought of as being "in a consensus round". Outside of this stage appears idle to the rest of the network.

### Validator's perspective

1. `PROPOSED` -- received new valid block proposal to execute from leader
2. `EXECUTED` -- block execution complete, prepared-transaction ready, appHash computed and sent to leader in ACK (or NACK)
3. `COMMITTED` -- committed the prepared tx on receipt of commit message from leader

Note that if the leader rollsback/aborts instead of requesting commit, we are back in the `COMMITTED` state for the *previous block* until we receive a new block proposal.

Optimization note: if the validator is slow, and a committed block announcement is received for the block that they are currently still executing (prior to sending their ACK), they may queue the commit even though their ACK will go ignored. This is an optimization because the alternative is to discard the (apparently early) commit message, and end up sending a pointless ACK that would be ignored or at best get a response containing an "ok to commit"/FF response. Worst case scenario would be the validator receives no response or a negative response to the ACK, and is forced to request the best block from peers as if it were in sync/catch-up. They will automatically be back in "consensus mode" when they receive the next block proposal that appears to be next from the perspective of that node's committed block index. Leader re-announcement will facilitate in the event of widespread network slowdown or other interruptions.

### Sentry's perspective

There is no awareness of the block production process, only committed block announcements. As long as the blocks have the leader's signature (or the message itself authenticates the leader as the source), the block is accepted and executed. Similarly during sync, but from blocks requested from peers rather in the announcement protocol.

### Recovery, Re-announcement, and Idempotence

A main reason why more fine grained stages are not necessary necessary is that messages can be re-sent without issue i.e. consensus messages should be idempotent. For instance, validators may re-announce their ACK/NACK, and leaders may re-announce their COMMIT.

In addition, robust handling of out-of-sequence messages simplifies recovery. For instance, if the leader proposes a new block, which validators begin executing, and the leader crashes before "precommitting" (creating the prepared transaction), and the leader starts up with no knowledge of the block they proposed, they are now back at the previous height where they now prepare and propose a new block. Validators have committed nothing and are able to rollback their prepared transactions for the abandoned proposal. Similarly, if the leader had precommitted, it may be rolled back on startup for the same outcome. If the crash was after leader commit, but before making the committed block announcement, leader startup can begin with a committed block announcement for the latest block. In the event of a clean restart, this would be an ignorable re-announcement of a known block to some or all of the validators.

Each protocol should be defined with this in mind. e.g. if leader receives an ACK for a block proposal that it does not recognize (either it is not in a consensus round or it has proposed a different block), a simple ABRT response can signal to the validator to rollback.

In any case where a validator cannot participate in the current round, they would simply wait for either the next committed block announcement (the round completed without them) or the next proposal (a new round was started). Requesting required blocks to catch up should be done until the validator is able to participate in consensus, a state that is trivially detected when a committed block announcement is for a block that is beyond the next in the validator's committed block index.
