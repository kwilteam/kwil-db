package dto

type AttributeType string

const (
	PRIMARY_KEY AttributeType = "PRIMARY_KEY"
	UNIQUE      AttributeType = "UNIQUE"
	NOT_NULL    AttributeType = "NOT_NULL"
	DEFAULT     AttributeType = "DEFAULT"
	MIN         AttributeType = "MIN"
	MAX         AttributeType = "MAX"
	MIN_LENGTH  AttributeType = "MIN_LENGTH"
	MAX_LENGTH  AttributeType = "MAX_LENGTH"
)

func (a AttributeType) String() string {
	return string(a)
}

func (a *AttributeType) IsValid() bool {
	return *a == PRIMARY_KEY || *a == UNIQUE || *a == NOT_NULL || *a == DEFAULT || *a == MIN || *a == MAX || *a == MIN_LENGTH || *a == MAX_LENGTH
}
