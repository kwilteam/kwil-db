package processor

import (
	"errors"
	st "kwil/x/deposits/structures"
	"kwil/x/logx"
	"math/big"
	"sync"
)

// Processor keeps in-memory information about balances, deposits, and withdrawals.

type Processor struct {
	wt    *st.WithdrawalTracker
	bals  map[string]*big.Int
	spent map[string]*big.Int
	log   logx.SugaredLogger
	mu    sync.Mutex
}

var ErrCantParseAmount = errors.New("can't parse amount")
var ErrInsufficientBalance = errors.New("insufficient balance")

// NewProcessor creates a new Processor instance.
func NewProcessor(l logx.Logger) *Processor {
	return &Processor{
		wt:    st.NewWithdrawalTracker(),
		bals:  map[string]*big.Int{},
		spent: map[string]*big.Int{},
		log:   l.Sugar().With("module", "processor"),
		mu:    sync.Mutex{},
	}
}

// Process Deposit increment the callers height by the amount
func (p *Processor) ProcessDeposit(d Deposit) error {
	curAmt := p.GetBalance(d.Caller())

	// parse the amount
	amt, ok := new(big.Int).SetString(d.Amount(), 10)
	if !ok {
		return ErrCantParseAmount
	}

	curAmt.Add(curAmt, amt)

	p.bals[d.Caller()] = curAmt
	return nil
}

// Process begin withdrawal subtracts the amount from the callers balance and puts the withdrawal
// in the withdrawal tracker
func (p *Processor) ProcessWithdrawalRequest(w WithdrawalRequest) error {
	// parse the amount

	withdrawAmt, ok := new(big.Int).SetString(w.Amount(), 10)
	if !ok {
		return ErrCantParseAmount
	}

	curAmt := p.GetBalance(w.Wallet())
	spentAmt := p.GetSpent(w.Wallet())

	if curAmt.String() == "0" && spentAmt.String() == "0" {
		// if both are nil, they have 0 funds so they can't withdraw
		return ErrInsufficientBalance
	}

	// now we need to check if the amount they are trying to withdraw is less than the amount they have spent
	newAmt := new(big.Int).Sub(curAmt, withdrawAmt)
	// check if newAmt is less than 0
	if newAmt.Cmp(big.NewInt(0)) == -1 {
		// if this is the case then we will just withdraw the amount they have left
		withdrawAmt = curAmt
		newAmt = big.NewInt(0)
	}

	p.setBalance(w.Wallet(), newAmt)
	p.setSpent(w.Wallet(), big.NewInt(0))

	wdrl := pendingWithdrawal{
		nonce:  w.Nonce(),
		amount: withdrawAmt,
		wallet: w.Wallet(),
		spent:  spentAmt,
		expiry: w.Expiration(),
	}

	p.wt.Insert(wdrl)

	return nil
}

// ProcessWithdrawalConfirmation removes the withdrawal from the withdrawal tracker
func (p *Processor) ProcessWithdrawalConfirmation(w WithdrawalConfirmation) {
	p.wt.RemoveByNonce(w.Nonce())
}

// ProcessFinalizedBlock removes all withdrawals that have expired and re-credits the account
func (p *Processor) ProcessFinalizedBlock(b FinalizedBlock) error {
	// pop all withdrawals that have expired
	expired := p.wt.PopExpired(b.Height())
	p.log.Infof("amount of expired withdrawals: %d", len(expired))
	for _, wdrl := range expired {
		// re-credit the account
		curAmt := p.GetBalance(wdrl.Item().Wallet())
		// turn withdrawal amount into a big int
		amt, ok := new(big.Int).SetString(wdrl.Item().Amount(), 10)
		if !ok {
			p.log.Warnf("can't parse deposit amount when processing unconfirmed withdrawal. not re-crediting account")
			p.logWithdrawal(wdrl)
			continue
		}
		curAmt.Add(curAmt, amt)

		// re-credit the spent amount
		curSpent := p.GetSpent(wdrl.Item().Wallet())
		// parsed the spent amount
		spentAmt, ok := new(big.Int).SetString(wdrl.Item().Spent(), 10)
		if !ok {
			p.log.Warnf("can't parse spent amount when processing unconfirmed withdrawal. not recrediting account")
			p.logWithdrawal(wdrl)
			continue
		}
		curSpent.Add(curSpent, spentAmt)
		p.setBalance(wdrl.Item().Wallet(), curAmt)
		p.setSpent(wdrl.Item().Wallet(), curSpent)
		p.log.Warnf("re-credited account for unconfirmed withdrawal")
		p.logWithdrawal(wdrl)
	}

	return nil
}

// ProcessSpend subtracts the amount from the callers balance
func (p *Processor) ProcessSpend(s Spend) error {
	// parse the amount
	amt, ok := new(big.Int).SetString(s.Amount(), 10)
	if !ok {
		return ErrCantParseAmount
	}

	curAmt := p.GetBalance(s.Caller())

	newAmt := curAmt.Sub(curAmt, amt) // do i need to set newAmt, or can I just use curAmt?
	// check if newAmt is less than 0
	cmp := newAmt.Cmp(big.NewInt(0))
	if cmp == -1 {
		return ErrInsufficientBalance
	}

	p.setBalance(s.Caller(), newAmt)
	p.setSpent(s.Caller(), amt)
	return nil
}

// GetBalance returns the callers balance
// if nil, return 0
func (p *Processor) GetBalance(addr string) *big.Int {
	bal := p.bals[addr]

	if bal == nil {
		return big.NewInt(0)
	}

	return bal
}

// GetSpent returns the amount spent by the caller
// if nil, return 0
func (p *Processor) GetSpent(addr string) *big.Int {
	spt := p.spent[addr]

	if spt == nil {
		return big.NewInt(0)
	}

	return spt
}

// setbalance sets the balance for a wallet.
// if the amount is 0 it should delete the key
func (pw *Processor) setBalance(addr string, amt *big.Int) {
	if amt.Cmp(big.NewInt(0)) == 0 {
		delete(pw.bals, addr)
		return
	}

	pw.bals[addr] = amt
}

// setspent sets the spent for a wallet
// if the amount is 0 it should delete the key
func (pw *Processor) setSpent(addr string, amt *big.Int) {
	if amt.Cmp(big.NewInt(0)) == 0 {
		delete(pw.spent, addr)
		return
	}

	pw.spent[addr] = amt
}

func (p *Processor) NonceExist(n string) bool {
	return p.wt.GetByNonce(n) != nil
}

// RunGC recreates the balances and spent maps.
// This is because golang maps are not garbage collected.
func (p *Processor) RunGC() {
	p.wt.RunGC() // we want to run this blocking
	// this outer function will likely be called non blocking, so p.wt.RunGC() has mutexs

	p.mu.Lock()
	defer p.mu.Unlock()

	nb := make(map[string]*big.Int)
	ns := make(map[string]*big.Int)

	for k, v := range p.bals {
		nb[k] = v
	}

	for k, v := range p.spent {
		ns[k] = v
	}

	p.bals = nb
	p.spent = ns
}

type pendingWithdrawal struct {
	amount *big.Int
	spent  *big.Int
	expiry int64
	wallet string
	nonce  string
}

func (pw pendingWithdrawal) Nonce() string {
	return pw.nonce
}

func (pw pendingWithdrawal) Expiration() int64 {
	return pw.expiry
}

func (pw pendingWithdrawal) Amount() string {
	return pw.amount.String()
}

func (pw pendingWithdrawal) Wallet() string {
	return pw.wallet
}

func (pw pendingWithdrawal) Spent() string {
	return pw.spent.String()
}

func (p *Processor) logWithdrawal(wdrl *st.Node) {
	p.log.Infof(`withdrawal:
	Wallet  | %s
	Deposit | %s
	Spent   | %s
	Nonce   | %s
	Expiry  | %s`, wdrl.Item().Wallet(), wdrl.Item().Amount(), wdrl.Item().Spent(), wdrl.Item().Nonce(), wdrl.Item().Expiration())
}
