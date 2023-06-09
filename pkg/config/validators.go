package config

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/kwilteam/kwil-db/pkg/utils"
)

type ApprovedValidators struct {
	Validators map[string]bool
	filePath   string
	mu         sync.RWMutex
}

func NewApprovedValidators(filePath string) *ApprovedValidators {
	return &ApprovedValidators{
		Validators: make(map[string]bool),
		filePath:   filePath,
		mu:         sync.RWMutex{},
	}
}

func (a *ApprovedValidators) AddValidator(address ed25519.PubKey) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	pubkey := string(address)
	if !a.Validators[pubkey] {
		// TODO: Check for the validity of the Validator pubkey
		a.Validators[pubkey] = true

		f, err := utils.OpenFile(a.filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND)
		if err != nil {
			fmt.Println("error opening ApprovedValidators file")
			return err
		}
		defer f.Close()

		_, err = f.Write([]byte(pubkey + "\n"))
		if err != nil {
			fmt.Println("error writing approved Validator to file")
			return err
		}
	}
	return nil
}

func (a *ApprovedValidators) IsValidator(address ed25519.PubKey) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Validators[string(address)]
}

func (a *ApprovedValidators) LoadOrCreateFile(filepath string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	data, err := utils.ReadOrCreateFile(filepath, os.O_CREATE|os.O_RDWR|os.O_APPEND)
	if err != nil {
		return err
	}

	validators := bytes.Split(data, []byte("\n"))
	for _, validator := range validators {
		a.Validators[string(validator)] = true
	}
	return nil
}
