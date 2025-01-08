package types

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"

	"github.com/kwilteam/kwil-db/core/crypto"
)

func init() {
	np := NetworkParameters{}
	setParamNames(np)

	// consistency checks to ensure changes to NetworkParameters have
	// corresponding changes in functions that switch on the ParamName.
	paramMap := np.ToMap()
	if len(paramMap) != numParams {
		panic("incorrect number of parameters defined in (NetworkParameters).ToMap")
	}
	if err := ValidateUpdates(paramMap); err != nil {
		panic("ValidateUpdateTypes: incorrect parameter types defined")
	}
	if err := MergeUpdates(&np, paramMap); err != nil {
		panic("MergeUpdates: incorrect parameter types defined")
	}
}

type PublicKey struct {
	crypto.PublicKey
}

func (pk PublicKey) String() string {
	// display key as hex and type as string
	return fmt.Sprintf("%s [%s]", hex.EncodeToString(pk.PublicKey.Bytes()), pk.Type())
}

type pubKeyJSON struct {
	Type string `json:"type"`
	Key  string `json:"key"`
}

func (pk PublicKey) MarshalJSON() ([]byte, error) {
	// Presently we reject nil public key.  We could write a null...
	if pk.PublicKey == nil {
		return nil, errors.New("nil public key")
		// return []byte("null"), nil
	}
	return json.Marshal(pubKeyJSON{
		Type: pk.Type().String(),
		Key:  hex.EncodeToString(pk.PublicKey.Bytes()),
	})
}

func (pk *PublicKey) UnmarshalJSON(b []byte) error {
	// To support null, we need to check for "null" string.
	// if bytes.Equal(b, []byte("null")) {
	// 	*pk = PublicKey{} // nil crypto.PublicKey
	// 	return nil
	// }
	var pkj pubKeyJSON
	if err := json.Unmarshal(b, &pkj); err != nil {
		return err
	}
	key, err := hex.DecodeString(pkj.Key)
	if err != nil {
		return err
	}
	kt, err := crypto.ParseKeyType(pkj.Type)
	if err != nil {
		return err
	}
	pubkey, err := crypto.UnmarshalPublicKey(key, kt)
	if err != nil {
		return err
	}
	*pk = PublicKey{pubkey}
	return nil
}

// NetworkParameters are network level configurations that can be
// evolved over the lifetime of a network.
type NetworkParameters struct {
	// Leader is the leader's public key. The leader must be in the current
	// validator set.
	Leader PublicKey `json:"leader"`

	// Validators set is logically also network parameters that evolve, but they
	// are tracked separately.

	// DBOwner is the owner of the database.
	// This should be either a public key or address.
	DBOwner string `json:"db_owner"`

	// MaxBlockSize is the maximum size of a block in bytes.
	MaxBlockSize int64 `json:"max_block_size"`
	// JoinExpiry is the number of blocks after which the validators
	// join request expires if not approved.
	JoinExpiry int64 `json:"join_expiry"`
	// VoteExpiry is the default number of blocks after which the validators
	// vote expires if not approved.
	VoteExpiry int64 `json:"vote_expiry"`
	// DisabledGasCosts dictates whether gas costs are disabled.
	DisabledGasCosts bool `json:"disabled_gas_costs"`
	// MaxVotesPerTx is the maximum number of votes that can be included in a
	// single transaction.
	MaxVotesPerTx int64 `json:"max_votes_per_tx"`

	MigrationStatus MigrationStatus `json:"migration_status"`
}

// ParamUpdates is the mechanism by which changes to network parameters are
// specified. Rather than a struct with pointer fields to indicate changes, we
// use a map to make updates easy to specify while keeping the NetworkParameters
// struct simple and easy to use without nil checks.
//
// This approach also makes update serialization more compact, only encoding
// data for updated fields.
//
// This approach however requires the definition of parameter names (ParamName)
// and code to assert type of the values. The parameter names are enumerated
// below, and their values are set using reflection during package initialization.
type ParamUpdates map[ParamName]any

type ParamName = string

// The ParamName values correspond to the fields of the NetworkParameters struct.
var (
	ParamNameLeader           ParamName
	ParamNameDBOwner          ParamName
	ParamNameMaxBlockSize     ParamName
	ParamNameJoinExpiry       ParamName
	ParamNameVoteExpiry       ParamName
	ParamNameDisabledGasCosts ParamName
	ParamNameMaxVotesPerTx    ParamName
	ParamNameMigrationStatus  ParamName
)

const numParams = 8

// setParamNames sets the ParamName constants based on the json tags of a struct
// (intended for NetworkParameters, but any for unit testing). This looks crazy,
// but it guarantees all fields of NetworkParameters have corresponding
// ParamName constants, and that the ParamName values are set based on the
// struct tags. It will panic if any field is missing a json tag or if there is
// no corresponding ParamName variable.
func setParamNames(np any) {
	var numFields int
	rt := reflect.TypeOf(np)
	for i := range rt.NumField() {
		field := rt.Field(i)
		fieldName := field.Name
		fieldTag := field.Tag.Get("json")
		if fieldTag == "" {
			panic(fmt.Sprintf("field %v lacks a json tag", field.Name))
		}
		switch fieldName {
		case "Leader":
			ParamNameLeader = fieldTag
		case "DBOwner":
			ParamNameDBOwner = fieldTag
		case "MaxBlockSize":
			ParamNameMaxBlockSize = fieldTag
		case "JoinExpiry":
			ParamNameJoinExpiry = fieldTag
		case "VoteExpiry":
			ParamNameVoteExpiry = fieldTag
		case "DisabledGasCosts":
			ParamNameDisabledGasCosts = fieldTag
		case "MaxVotesPerTx":
			ParamNameMaxVotesPerTx = fieldTag
		case "MigrationStatus":
			ParamNameMigrationStatus = fieldTag
		default:
			panic(fmt.Sprintf("unknown field %v", fieldName))
		}
		numFields++
	}
	if numFields != numParams {
		panic("not all fields have corresponding ParamName constants")
	}
}

func MergeUpdates(np *NetworkParameters, updates ParamUpdates) (err error) {
	// if err = ValidateUpdateTypes(updates); err != nil {
	// 	return err
	// }
	defer func() { // failed type assertion, bug in config package
		if r := recover(); r != nil {
			err = fmt.Errorf("invalid update: %v", r) // could also be nil *NetworkParameters
		}
	}()

	for paramName, update := range updates {
		switch paramName {
		case ParamNameLeader:
			switch key := update.(type) {
			case PublicKey:
				np.Leader = key
			case crypto.PublicKey:
				np.Leader = PublicKey{key}
			default:
				return fmt.Errorf("invalid type for leader: %T", update)
			}
		case ParamNameDBOwner:
			np.DBOwner = update.(string)
		case ParamNameMaxBlockSize:
			np.MaxBlockSize = update.(int64)
		case ParamNameJoinExpiry:
			np.JoinExpiry = update.(int64)
		case ParamNameVoteExpiry:
			np.VoteExpiry = update.(int64)
		case ParamNameDisabledGasCosts:
			np.DisabledGasCosts = update.(bool)
		case ParamNameMaxVotesPerTx:
			np.MaxVotesPerTx = update.(int64)
		case ParamNameMigrationStatus:
			np.MigrationStatus = update.(MigrationStatus)
		default:
			return fmt.Errorf("unknown field %v", paramName)
		}
	}
	return nil
}

func ValidateUpdates(pu ParamUpdates) error {
	np := NetworkParameters{}
	return MergeUpdates(&np, pu)
}

func (pu ParamUpdates) Merge(other ParamUpdates) {
	for k, v := range other {
		pu[k] = v
	}
}

func (pu ParamUpdates) Equals(other ParamUpdates) bool {
	if len(pu) != len(other) {
		return false
	}
	if len(pu) == 0 && len(other) == 0 {
		return true // consider nil and empty equal
	}
	// Same length and both non-nil => check for equality
	np0 := &NetworkParameters{}
	MergeUpdates(np0, pu)
	np1 := &NetworkParameters{}
	MergeUpdates(np1, other)
	return np0.Equals(np1)
}

func (pu ParamUpdates) String() string {
	bts, err := json.Marshal(pu)
	if err != nil {
		return "<invalid>"
	}
	return string(bts)
}

// Bytes returns the serialization of the updates. This will panic if there are
// invalid update keys. Use ValidateUpdates first to ensure they are valid.
func (pu ParamUpdates) Bytes() []byte {
	bts, err := pu.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return bts
}

// MarshalBinary encodes ParamUpdates into a binary format.
func (pu ParamUpdates) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}

	// Write a version const
	const version = 0
	if err := binary.Write(buf, binary.LittleEndian, uint16(version)); err != nil {
		return nil, err
	}

	// Serialize the number of updates
	if err := binary.Write(buf, binary.LittleEndian, uint16(len(pu))); err != nil {
		return nil, err
	}

	// Serialize each update deterministically
	keys := make([]string, 0, len(pu))
	for key := range pu {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		// Write the parameter name length and name
		if err := binary.Write(buf, binary.LittleEndian, uint16(len(key))); err != nil {
			return nil, err
		}
		if _, err := buf.Write([]byte(key)); err != nil {
			return nil, err
		}

		// Serialize the value based on the type
		value := pu[key]
		switch key {
		case ParamNameLeader:
			var pk PublicKey
			switch val := value.(type) {
			case PublicKey:
				pk = val
			case crypto.PublicKey:
				pk = PublicKey{val}
			default:
				return nil, fmt.Errorf("invalid type for %s", key)
			}

			bts := crypto.WireEncodePublicKey(pk)
			if err := binary.Write(buf, binary.LittleEndian, uint16(len(bts))); err != nil {
				return nil, err
			}
			if _, err := buf.Write(bts); err != nil {
				return nil, err
			}
		case ParamNameDBOwner:
			if val, ok := value.(string); ok {
				if err := binary.Write(buf, binary.LittleEndian, uint16(len(val))); err != nil {
					return nil, err
				}
				if _, err := buf.Write([]byte(val)); err != nil {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("invalid type for %s", key)
			}
		case ParamNameMaxBlockSize, ParamNameJoinExpiry, ParamNameVoteExpiry, ParamNameMaxVotesPerTx:
			if val, ok := value.(int64); ok {
				if err := binary.Write(buf, binary.LittleEndian, val); err != nil {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("invalid type for %s", key)
			}
		case ParamNameDisabledGasCosts:
			if val, ok := value.(bool); ok {
				var boolInt uint8
				if val {
					boolInt = 1
				}
				if err := binary.Write(buf, binary.LittleEndian, boolInt); err != nil {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("invalid type for %s", key)
			}
		case ParamNameMigrationStatus:
			if val, ok := value.(MigrationStatus); ok {
				statusBts := []byte(val)
				if err := binary.Write(buf, binary.LittleEndian, uint16(len(statusBts))); err != nil {
					return nil, err
				}
				if _, err := buf.Write(statusBts); err != nil {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("invalid type for %s", key)
			}
		default:
			return nil, fmt.Errorf("unknown parameter name: %s", key)
		}
	}

	return buf.Bytes(), nil
}

// UnmarshalBinary decodes a binary format into ParamUpdates.
func (pu *ParamUpdates) UnmarshalBinary(data []byte) error {
	buf := bytes.NewReader(data)

	// Check the version
	var version uint16
	if err := binary.Read(buf, binary.LittleEndian, &version); err != nil {
		return err
	}
	if version != 0 {
		return fmt.Errorf("unsupported version: %d", version)
	}
	// Different future versions will support different param names and possibly
	// types. Presently, the following code is effectively unmarshalV0().

	// Read the number of updates
	var numUpdates uint16 // 65535 is more than enough parameters
	if err := binary.Read(buf, binary.LittleEndian, &numUpdates); err != nil {
		return err
	}

	updates := make(ParamUpdates, numUpdates)
	for range numUpdates {
		// Read the parameter name
		var nameLen uint16
		if err := binary.Read(buf, binary.LittleEndian, &nameLen); err != nil {
			return err
		}
		nameBytes := make([]byte, nameLen)
		if _, err := buf.Read(nameBytes); err != nil {
			return err
		}
		paramName := ParamName(nameBytes)

		// Deserialize the value based on the parameter name
		switch paramName {
		case ParamNameLeader:
			var length uint16
			if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
				return err
			}
			val := make([]byte, length)
			if _, err := buf.Read(val); err != nil {
				return err
			}
			pubkey, err := crypto.WireDecodePubKey(val)
			if err != nil {
				return err
			}
			updates[paramName] = PublicKey{pubkey}
		case ParamNameDBOwner:
			var length uint16
			if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
				return err
			}
			val := make([]byte, length)
			if _, err := buf.Read(val); err != nil {
				return err
			}
			updates[paramName] = string(val)
		case ParamNameMaxBlockSize, ParamNameJoinExpiry, ParamNameVoteExpiry, ParamNameMaxVotesPerTx:
			var val int64
			if err := binary.Read(buf, binary.LittleEndian, &val); err != nil {
				return err
			}
			updates[paramName] = val
		case ParamNameDisabledGasCosts:
			var val uint8
			if err := binary.Read(buf, binary.LittleEndian, &val); err != nil {
				return err
			}
			updates[paramName] = val == 1
		case ParamNameMigrationStatus:
			var length uint16
			if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
				return err
			}
			val := make([]byte, length)
			if _, err := buf.Read(val); err != nil {
				return err
			}
			updates[paramName] = MigrationStatus(val)
		default:
			return fmt.Errorf("unknown parameter name: %s", paramName)
		}
	}

	*pu = updates

	return nil
}

func (np NetworkParameters) ToMap() map[ParamName]any {
	// Create a map using ParamNames as keys.
	return map[ParamName]any{
		ParamNameLeader:           np.Leader,
		ParamNameDBOwner:          np.DBOwner,
		ParamNameMaxBlockSize:     np.MaxBlockSize,
		ParamNameJoinExpiry:       np.JoinExpiry,
		ParamNameVoteExpiry:       np.VoteExpiry,
		ParamNameDisabledGasCosts: np.DisabledGasCosts,
		ParamNameMaxVotesPerTx:    np.MaxVotesPerTx,
		ParamNameMigrationStatus:  np.MigrationStatus,
	}
}

func (np NetworkParameters) MarshalBinary() ([]byte, error) {
	return json.Marshal(np)
}

func (np *NetworkParameters) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, np)
}

func (np *NetworkParameters) Clone() *NetworkParameters {
	paramsCopy := *np
	return &paramsCopy
}

func (np *NetworkParameters) Equals(other *NetworkParameters) bool {
	if np == nil && other == nil {
		return true
	}
	if np == nil || other == nil {
		return false
	}
	var sameLeader bool
	if np.Leader.PublicKey == nil && other.Leader.PublicKey == nil {
		sameLeader = true
	} else if np.Leader.PublicKey != nil && other.Leader.PublicKey != nil {
		sameLeader = np.Leader.Equals(other.Leader)
	}
	return sameLeader &&
		np.DBOwner == other.DBOwner &&
		np.MaxBlockSize == other.MaxBlockSize &&
		np.JoinExpiry == other.JoinExpiry &&
		np.VoteExpiry == other.VoteExpiry &&
		np.DisabledGasCosts == other.DisabledGasCosts &&
		np.MaxVotesPerTx == other.MaxVotesPerTx &&
		np.MigrationStatus == other.MigrationStatus
}

func (np *NetworkParameters) SanityChecks() error {
	// Leader shouldn't be empty
	if np.Leader.PublicKey == nil || len(np.Leader.Bytes()) == 0 || !np.Leader.Type().Valid() {
		return errors.New("leader should not be empty")
	}
	// DBOwner shouldn't be empty
	if len(np.DBOwner) == 0 {
		return errors.New("db owner should not be empty")
	}

	// Vote expiry shouldn't be 0
	if np.VoteExpiry == 0 {
		return errors.New("vote expiry should be greater than 0")
	}

	// MaxVotesPerTx shouldn't be 0
	if np.MaxVotesPerTx == 0 {
		return errors.New("max votes per tx should be greater than 0")
	}

	// join expiry shouldn't be 0
	if np.JoinExpiry == 0 {
		return errors.New("join expiry should be greater than 0")
	}

	// Block params
	if np.MaxBlockSize == 0 {
		return errors.New("max bytes should be greater than 0")
	}

	if !np.MigrationStatus.Valid() {
		return fmt.Errorf("migration status is invalid: %q", np.MigrationStatus)
	}

	return nil
}

func (np NetworkParameters) String() string {
	return fmt.Sprintf(`Network Parameters:
	Leader: %s
	DB Owner: %s
	Max Block Size: %d
	Join Expiry: %d
	Vote Expiry: %d
	Disabled Gas Costs: %t
	Max Votes Per Tx: %d
	Migration Status: %s`,
		np.Leader, np.DBOwner, np.MaxBlockSize, np.JoinExpiry,
		np.VoteExpiry, np.DisabledGasCosts, np.MaxVotesPerTx, np.MigrationStatus)
}
