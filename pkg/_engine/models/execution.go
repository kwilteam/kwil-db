package models

type ActionExecution struct {
	Action string              `json:"action"`
	DBID   string              `json:"dbid"`
	Params []map[string][]byte `json:"params"`
}

const (
	Max_Text_Length = 1024
)

type ActionInstance map[string][]byte

// Ok will check that all values are acceptable.
// For example, if a string is too long, it will return false.
func (a ActionInstance) Ok() bool {
	for _, v := range a {
		if len(v) > Max_Text_Length+1 { // we add 1 to account for the type byte
			return false
		}
	}

	return true
}
