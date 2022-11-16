package syntax

const (
	DiagInvalidLine          = "Invalid line"
	DiagUnexpectedToken      = "Unexpected token"
	DiagUnexpectedBlockStart = "Unexpected block start"

	// Directives
	DiagExpectedDirectiveOrBlock     = "Expected directive or block"
	DiagExpectedDirective            = "Expected directive"
	DiagInvalidDirectiveName         = "Invalid directive name"
	DiagNoKeyBeforeAssignment        = "No key before assignment"
	DiagMissingDirectiveValue        = "Missing directive value"
	DiagMissingNewlineAfterDirective = "Missing newline after directive"

	// Models
	DiagExpectedModelDefinition = "Expected model definition"
	DiagInvalidModelDefinition  = "Invalid model definition"
	DiagInvalidEnumName         = "Invalid enum name"
	DiagInvalidEnumDefinition   = "Invalid enum definition"
	DiagInvalidEnumValue        = "Invalid enum value"
	DiagInvalidFieldDefinition  = "Invalid field definition"

	// Blocks
	DiagExpectedBlockDefinition    = "Expected block definition"
	DiagExpectedBlockClose         = "Expected block close"
	DiagInvalidBlockDefinition     = "Invalid block definition"
	DiagBlockInvalidModifierTarget = "Invalid modifier target"
	DiagMissingNewlineAfterBlock   = "Missing newline after block definition"

	// Block body
	DiagInvalidBlockBody       = "Invalid block body"
	DiagInvalidDeclarationName = "Invalid declaration name"
	DiagInvalidDeclaration     = "Invalid declaration"
	DiagInvalidAnnotation      = "Invalid annotation"

	// Block labels
	DiagInvalidLabelName     = "Invalid label name"
	DiagInvalidLabel         = "Invalid label"
	DiagInvalidTypeName      = "Invalid type name"
	DiagInvalidTypeSpecifier = "Invalid type specifier"

	DiagInvalidFunctionName       = "Invalid function name"
	DiagInvalidKeyValue           = "Invalid key-value"
	DiagInvalidExpression         = "Invalid expression"
	DiagInvalidPropertyName       = "Invalid property name"
	DiagInvalidPropertyAssignment = "Invalid property assignment"
	DiagUnterminatedObject        = "Unterminated object"
	DiagMissingSeparator          = "Missing separator"
	DiagExpectedNumberLit         = "Expected number literal"
	DiagMissingParenthesis        = "Missing parenthesis"
	DiagInvalidArgument           = "Invalid argument"
	DiagUnterminatedList          = "Unterminated list"
)

const (
	DiagFloatingDocComment = "Floating doc comment"
)
