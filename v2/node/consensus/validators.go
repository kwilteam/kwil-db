package consensus

// How to handle validatorset updates???
//
/*
	Transaction Model:
	- Transaction execution can directly affect the validator set
	- ValidatorSetHash can dictate the validator set updates.
	- Do we need leader to send the updates in the header in this model?? probably not. as anyone doing a statesync or blocksync can get the to the current state by replaying the blocks. and the transactions can directly affect the validator set.

	P2P Model:
	- Instead of transactions, p2p messages can be used to update the validator set (join, leave, remove, etc.)
	- The leader can send the validator set updates to the validators in the block proposal.
	- Validators can reject the block if the validator set updates are not valid. On what basis can they reject the block??
		- maybe leader has to send the validators signature for the given validator set update.
	- ValidatorsHash can be used to verify the validator set for a given round of consensus.


	similarly consensus param updates:
*/

