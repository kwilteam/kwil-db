package abci

type TxCode uint32

const (
	CodeOk            TxCode = 0
	CodeEncodingError TxCode = 1
	CodeUnknownError  TxCode = 2 // for now it's for all non-encoding error
)

func (c TxCode) Uint32() uint32 {
	return uint32(c)
}
