package rules

// Code is the name for the field constraint.
type Code string

const (
	Required  Code = "REQUIRED"
	NotNull   Code = "NOT_NULL"
	Boolean   Code = "BOOLEAN"
	Number    Code = "NUMBER"
	String    Code = "STRING"
	Length    Code = "LENGTH"
	MinLength Code = "MIN_LENGTH"
	MaxLength Code = "MAX_LENGTH"
	Pattern   Code = "PATTERN"
	Email     Code = "EMAIL"
)
