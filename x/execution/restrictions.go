package execution

// max name lengths
const (
	MAX_OWNER_NAME_LENGTH     = 44 // ETH addresses are 42, Solana are 44
	MAX_COLUMN_NAME_LENGTH    = 32 // 59 is max for Postgres
	MAX_TABLE_NAME_LENGTH     = 32 // 63 is max for Postgres
	MAX_INDEX_NAME_LENGTH     = 32 // 63 is max for Postgres
	MAX_ROLE_NAME_LENGTH      = 32 // Kwil roles != Postgres roles, so this is arbitrary
	MAX_DB_NAME_LENGTH        = 32 // 63 is nax schema name, this gets included in the schema name
	MAX_QUERY_NAME_LENGTH     = 32 // arbitrary
	MAX_ATTRIBUTE_NAME_LENGTH = 32 // arbitrary
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
	MAX_INDEX_COLUMNS = 3
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
