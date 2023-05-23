package validation

// max name lengths
const (
	MAX_OWNER_NAME_LENGTH = 44 // ETH addresses are 42, Solana are 44
	MAX_IDENT_NAME_LENGTH = 32
)

// table restrictions
const (
	MAX_TABLE_COUNT           = 100
	MAX_COLUMNS_PER_TABLE     = 50 // per table
	MAX_ATTRIBUTES_PER_COLUMN = 5  // per column
)

// index restrictions
const (
	MAX_INDEX_COUNT   = 100
	MAX_INDEX_COLUMNS = 5
)

// query restrictions
const (
	MAX_QUERY_COUNT     = 400
	MAX_PARAM_PER_QUERY = 50
	MAX_WHERE_PER_QUERY = 3
)

// role restrictions
const (
	MAX_ROLE_COUNT = 50
)

// wallets
const (
	MIN_WALLET_LENGTH = 42
	MAX_WALLET_LENGTH = 44
)
