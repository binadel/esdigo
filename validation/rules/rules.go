package rules

// Code is the name for the field constraint.
type Code string

const (
	Required Code = "REQUIRED"
	NotNull  Code = "NOT_NULL"
	Boolean  Code = "BOOLEAN"
	Number   Code = "NUMBER"
	String   Code = "STRING"
	Email    Code = "EMAIL"
)
