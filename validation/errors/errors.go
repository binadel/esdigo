package errors

import "github.com/binadel/esdigo/validation/rules"

var (
	Required = &BasicError{
		Code:    rules.Required,
		Message: "field is required",
	}

	NotNull = &BasicError{
		Code:    rules.NotNull,
		Message: "value must be not null",
	}

	InvalidBoolean = &BasicError{
		Code:    rules.Boolean,
		Message: "field must be a valid boolean value",
	}

	InvalidNumber = &BasicError{
		Code:    rules.Number,
		Message: "field must be a valid number value",
	}

	InvalidString = &BasicError{
		Code:    rules.String,
		Message: "field must be a valid string value",
	}
)
