package balances

func (a *AccountStore) SetHeight(chainCode int32, height int64) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.setChainHeight(chainCode, height)
}

func (a *AccountStore) GetHeight(chainCode int32) (int64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.getChainHeight(chainCode)
}

func (a *AccountStore) CreateChain(chainCode int32, height int64) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.createChain(chainCode, height)
}

func (a *AccountStore) ChainExists(chainCode int32) (bool, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.chainExists(chainCode)
}
