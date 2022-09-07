package types

import (
	"crypto"
	"strings"
)

/*
CreateDatabaseMsg is the message sent to create a database

To generate ID: SHA384

  - From 		| Guarantees Uniquess

  - Name 		| Guarantees Uniquess

  - Fee 		| Verifies sender's consent to pay fee

    We don't have to worry much about collisions since databases are owned by a single wallet.
    Also not sure why it's auto-formatting this so weirdly.
*/
type CreateDatabaseMsg struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	DBType    string `json:"type"`
	Fee       string `json:"fee"`
	Operation byte   `json:"operation"`
	Crud      byte   `json:"crud"`
	From      string `json:"from"`
	Signature string `json:"signature"`
}

func (m *CreateDatabaseMsg) regenID() string {
	sb := strings.Builder{}
	sb.WriteString(m.From)
	sb.WriteString(m.Name)
	sb.WriteString(m.Fee)

	return string(sha384([]byte(sb.String())))
}

/*
	CreateDatabaseResponse is the response from creating a database

	To generate ID: SHA384
		- Owner 		| Guarantees uniqueness from other dbs
		- Name 			| Guarantees uniqueness from other dbs
		- Fee 			| Verifies sender's consent to withdraw fee
		- Operation 	| Verifies sender's consent to perform the operation
		- Crud 			| Verifies sender's consent to perform the operation
		- Instruction 	| Verifies sender's consent to perform the operation
		- From 			| Ensures that malicious actors can't cause collisions by matching all other fields
		- Nonce 		| Ensures that the same client can logically send the same message without causing a collision
*/

type AlterDatabaseMsg struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Owner        string `json:"owner"`
	Fee          string `json:"fee"`
	Operation    byte   `json:"operation"`
	Crud         byte   `json:"crud"`
	Instructions string `json:"instructions"`
	From         string `json:"from"`
	Nonce        string `json:"nonce"`
	Signature    string `json:"signature"`
}

func (m *AlterDatabaseMsg) regenID() string {
	sb := strings.Builder{}
	sb.WriteString(m.Owner)
	sb.WriteString(m.Name)
	sb.WriteString(m.Fee)
	sb.WriteByte(m.Operation)
	sb.WriteByte(m.Crud)
	sb.WriteString(m.Instructions)
	sb.WriteString(m.From)
	sb.WriteString(m.Nonce)

	return string(sha384([]byte(sb.String())))
}

func sha384(data []byte) []byte {
	return crypto.SHA384.New().Sum(data)
}

func (m *CreateDatabaseMsg) CheckID() bool {
	return m.ID == m.regenID()
}

func (m *AlterDatabaseMsg) CheckID() bool {
	return m.ID == m.regenID()
}
