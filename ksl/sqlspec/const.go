package sqlspec

// Standard column types (and their aliases) as defined in
// PostgreSQL codebase/website.
const (
	TypeBit     = "bit"
	TypeBitVar  = "bit varying"
	TypeBoolean = "boolean"
	TypeBool    = "bool" // boolean.
	TypeBytea   = "bytea"

	TypeCharacter = "character"
	TypeChar      = "char" // character
	TypeCharVar   = "character varying"
	TypeVarChar   = "varchar" // character varying
	TypeText      = "text"

	TypeSmallInt = "smallint"
	TypeInteger  = "integer"
	TypeBigInt   = "bigint"
	TypeInt      = "int"  // integer.
	TypeInt2     = "int2" // smallint.
	TypeInt4     = "int4" // integer.
	TypeInt8     = "int8" // bigint.

	TypeCIDR     = "cidr"
	TypeInet     = "inet"
	TypeMACAddr  = "macaddr"
	TypeMACAddr8 = "macaddr8"

	TypeCircle  = "circle"
	TypeLine    = "line"
	TypeLseg    = "lseg"
	TypeBox     = "box"
	TypePath    = "path"
	TypePolygon = "polygon"
	TypePoint   = "point"

	TypeDate          = "date"
	TypeTime          = "time"   // time without time zone
	TypeTimeTZ        = "timetz" // time with time zone
	TypeTimeWTZ       = "time with time zone"
	TypeTimeWOTZ      = "time without time zone"
	TypeTimestamp     = "timestamp" // timestamp without time zone
	TypeTimestampTZ   = "timestamptz"
	TypeTimestampWTZ  = "timestamp with time zone"
	TypeTimestampWOTZ = "timestamp without time zone"

	TypeDouble = "double precision"
	TypeReal   = "real"
	TypeFloat8 = "float8" // double precision
	TypeFloat4 = "float4" // real
	TypeFloat  = "float"  // float(p).

	TypeNumeric = "numeric"
	TypeDecimal = "decimal" // numeric

	TypeSmallSerial = "smallserial" // smallint with auto_increment.
	TypeSerial      = "serial"      // integer with auto_increment.
	TypeBigSerial   = "bigserial"   // bigint with auto_increment.
	TypeSerial2     = "serial2"     // smallserial
	TypeSerial4     = "serial4"     // serial
	TypeSerial8     = "serial8"     // bigserial

	TypeArray       = "array"
	TypeXML         = "xml"
	TypeJSON        = "json"
	TypeJSONB       = "jsonb"
	TypeUUID        = "uuid"
	TypeMoney       = "money"
	TypeInterval    = "interval"
	TypeUserDefined = "user-defined"
)

// List of supported index types.
const (
	IndexTypeBTree = "BTREE"
	IndexTypeHash  = "HASH"
	IndexTypeGIN   = "GIN"
	IndexTypeGiST  = "GIST"
	IndexTypeBRIN  = "BRIN"
)

const (
	GeneratedTypeAlways    = "ALWAYS"
	GeneratedTypeByDefault = "BY_DEFAULT" // BY DEFAULT.
)

// List of PARTITION KEY types.
const (
	PartitionTypeRange = "RANGE"
	PartitionTypeList  = "LIST"
	PartitionTypeHash  = "HASH"
)

const DefaultTimePrecision = 6
const DefaultPagePerRange = 128

const (
	KwilManagementSchema = "kwil"
	KwilRoleTable        = "roles"
	KwilQueryTable       = "queries"
	KwilRoleQueriesTable = "role_queries"
)
