package types

import "github.com/kwilteam/kwil-db/pkg/serialize/rlp"

type ActionExecution struct {
	DBID      string
	Action    string
	Arguments [][]string
}

func (a *ActionExecution) Bytes() ([]byte, error) {
	return rlp.Encode(a)
}

func (s *ActionExecution) FromBytes(b []byte) error {
	res, err := rlp.Decode[ActionExecution](b)
	if err != nil {
		return err
	}

	*s = *res
	return nil
}

type ActionCall struct {
	DBID      string
	Action    string
	Arguments []string
}

func (a *ActionCall) Bytes() ([]byte, error) {
	return rlp.Encode(a)
}

func (s *ActionCall) FromBytes(b []byte) error {
	res, err := rlp.Decode[ActionCall](b)
	if err != nil {
		return err
	}

	*s = *res
	return nil
}
