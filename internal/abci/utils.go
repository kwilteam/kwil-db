package abci

import (
	"crypto/sha256"

	"github.com/kwilteam/kwil-db/common/chain"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/extensions/consensus"
)

type txOut struct {
	tx   *transactions.Transaction
	hash [32]byte
}

// groupTransactions groups the transactions by sender.
func groupTxsBySender(txns [][]byte) (map[string][]*txOut, error) {
	grouped := make(map[string][]*txOut)
	for _, tx := range txns {
		t := &transactions.Transaction{}
		err := t.UnmarshalBinary(tx)
		if err != nil {
			return nil, err
		}
		key := string(t.Sender)
		grouped[key] = append(grouped[key], &txOut{
			tx:   t,
			hash: sha256.Sum256(tx),
		})
	}
	return grouped, nil
}

// nonceList is for debugging
func nonceList(txns []*txOut) []uint64 {
	nonces := make([]uint64, len(txns))
	for i, tx := range txns {
		nonces[i] = tx.tx.Body.Nonce
	}
	return nonces
}

func updateConsensusParams(p *chain.ConsensusParams, up *consensus.ParamUpdates) {
	if up.Block != nil { // gas and bytes can be independently set / unset
		if maxGas := up.Block.MaxGas; maxGas != 0 {
			p.Block.MaxGas = maxGas
		}
		if maxBts := up.Block.MaxBytes; maxBts != 0 {
			p.Block.MaxBytes = maxBts
			p.Block.AbciBlockSizeHandling = up.Block.AbciBlockSizeHandling // kwil-specific
		}
	}
	if up.Evidence != nil { // if set, expect all set
		p.Evidence.MaxAgeDuration = up.Evidence.MaxAgeDuration
		p.Evidence.MaxAgeNumBlocks = up.Evidence.MaxAgeNumBlocks
		p.Evidence.MaxBytes = up.Evidence.MaxBytes
	}
	if up.Version != nil { // if set, expect all set
		p.Version.App = up.Version.App
	}
	if up.Validator != nil {
		if pkt := up.Validator.PubKeyTypes; len(pkt) > 0 {
			p.Validator.PubKeyTypes = pkt
		}
		if exp := up.Validator.JoinExpiry; exp != 0 { // kwil-specific
			p.Validator.JoinExpiry = exp
		}
	}
	if up.Votes != nil { // entirely kwil-specific
		if exp := up.Votes.VoteExpiry; exp != 0 {
			p.Votes.VoteExpiry = exp
		}
	}
	if up.ABCI != nil {
		// Disabling this after it was enabled is impossible under cometbft
		// design. Only define an update that disables it if it had never
		// reached a previously configured enable height.
		p.ABCI.VoteExtensionsEnableHeight = up.ABCI.VoteExtensionsEnableHeight
	}
}
