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

	Length = BasicError{
		Code:    rules.Length,
		Message: "value length must be equal to {{length}}",
	}

	MinLength = BasicError{
		Code:    rules.MinLength,
		Message: "value length must be at least {{minLength}}",
	}

	MaxLength = BasicError{
		Code:    rules.MaxLength,
		Message: "value length must be at most {{maxLength}}",
	}

	Pattern = &BasicError{
		Code:    rules.Pattern,
		Message: "value does not match the required pattern",
	}

	InvalidEmail = &BasicError{
		Code:    rules.Email,
		Message: "value must be a valid email",
	}

	InvalidUri = &BasicError{
		Code:    rules.Uri,
		Message: "value must be a valid URI",
	}

	InvalidUuid = &BasicError{
		Code:    rules.Uuid,
		Message: "value must be a valid UUID",
	}

	UuidVersion = BasicError{
		Code:    rules.UuidVersion,
		Message: "value must be a valid UUID version {{version}}",
	}
)
