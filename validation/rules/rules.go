package rules

// Code is the name for the field constraint.
type Code string

const (
	Required    Code = "REQUIRED"
	NotNull     Code = "NOT_NULL"
	Boolean     Code = "BOOLEAN"
	Number      Code = "NUMBER"
	String      Code = "STRING"
	Array       Code = "ARRAY"
	Object      Code = "OBJECT"
	Length      Code = "LENGTH"
	MinLength   Code = "MIN_LENGTH"
	MaxLength   Code = "MAX_LENGTH"
	Pattern     Code = "PATTERN"
	Regex       Code = "REGEX"
	Date        Code = "DATE"
	Time        Code = "TIME"
	DateTime    Code = "DATE_TIME"
	Duration    Code = "DURATION"
	Email       Code = "EMAIL"
	IP          Code = "IP"
	IPv4        Code = "IPv4"
	IPv6        Code = "IPv6"
	Uri         Code = "URI"
	Uuid        Code = "UUID"
	UuidVersion Code = "UUID_VERSION"
	ExactItems  Code = "EXACT_ITEMS"
	MinItems    Code = "MIN_ITEMS"
	MaxItems    Code = "MAX_ITEMS"
)
