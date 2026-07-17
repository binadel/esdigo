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

	NotInteger = &BasicError{
		Code:    rules.Integer,
		Message: "field must be an integer",
	}

	OutOfRange = &BasicError{
		Code:    rules.OutOfRange,
		Message: "value is out of the allowed range",
	}

	InvalidString = &BasicError{
		Code:    rules.String,
		Message: "field must be a valid string value",
	}

	InvalidArray = &BasicError{
		Code:    rules.Array,
		Message: "field must be a valid array value",
	}

	InvalidObject = &BasicError{
		Code:    rules.Object,
		Message: "field must be a valid object value",
	}

	Length = BasicError{
		Code:    rules.Length,
		Message: "value must have the exact length",
	}

	MinLength = BasicError{
		Code:    rules.MinLength,
		Message: "value must be at least the minimum length",
	}

	MaxLength = BasicError{
		Code:    rules.MaxLength,
		Message: "value must be at most the maximum length",
	}

	Pattern = &BasicError{
		Code:    rules.Pattern,
		Message: "value does not match the required pattern",
	}

	InvalidRegex = &BasicError{
		Code:    rules.Regex,
		Message: "value must a valid regular expression",
	}

	Date = &BasicError{
		Code:    rules.Date,
		Message: "value must be a valid date",
	}

	Time = &BasicError{
		Code:    rules.Time,
		Message: "value must be a valid time",
	}

	DateTime = &BasicError{
		Code:    rules.DateTime,
		Message: "value must be a valid date time",
	}

	Duration = &BasicError{
		Code:    rules.Duration,
		Message: "value must be a valid duration",
	}

	InvalidEmail = &BasicError{
		Code:    rules.Email,
		Message: "value must be a valid email",
	}

	InvalidIP = &BasicError{
		Code:    rules.IP,
		Message: "value must be a valid IP address",
	}

	InvalidIPv4 = &BasicError{
		Code:    rules.IPv4,
		Message: "value must be a valid IPv4 address",
	}

	InvalidIPv6 = &BasicError{
		Code:    rules.IPv6,
		Message: "value must be a valid IPv6 address",
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
		Message: "value must be a UUID of the required version",
	}

	ExactItems = BasicError{
		Code:    rules.ExactItems,
		Message: "value must contain exactly the required number of items",
	}

	MinItems = BasicError{
		Code:    rules.MinItems,
		Message: "value must contain at least the minimum number of items",
	}

	MaxItems = BasicError{
		Code:    rules.MaxItems,
		Message: "value must contain at most the maximum number of items",
	}

	MinProperties = BasicError{
		Code:    rules.MinProperties,
		Message: "object must have at least the minimum number of properties",
	}

	MaxProperties = BasicError{
		Code:    rules.MaxProperties,
		Message: "object must have at most the maximum number of properties",
	}

	Minimum = BasicError{
		Code:    rules.Minimum,
		Message: "value must be at least the minimum",
	}

	Maximum = BasicError{
		Code:    rules.Maximum,
		Message: "value must be at most the maximum",
	}

	ExclusiveMinimum = BasicError{
		Code:    rules.ExclusiveMinimum,
		Message: "value must be greater than the exclusive minimum",
	}

	ExclusiveMaximum = BasicError{
		Code:    rules.ExclusiveMaximum,
		Message: "value must be less than the exclusive maximum",
	}

	MultipleOf = BasicError{
		Code:    rules.MultipleOf,
		Message: "value must be a multiple of the factor",
	}

	Enum = BasicError{
		Code:    rules.Enum,
		Message: "value must be one of the allowed values",
	}

	Const = BasicError{
		Code:    rules.Const,
		Message: "value must equal the required constant",
	}

	UniqueItems = &BasicError{
		Code:    rules.UniqueItems,
		Message: "array items must be unique",
	}

	InvalidHostname = &BasicError{
		Code:    rules.Hostname,
		Message: "value must be a valid hostname",
	}

	InvalidUriReference = &BasicError{
		Code:    rules.UriReference,
		Message: "value must be a valid URI reference",
	}

	InvalidJsonPointer = &BasicError{
		Code:    rules.JsonPointer,
		Message: "value must be a valid JSON pointer",
	}
)
