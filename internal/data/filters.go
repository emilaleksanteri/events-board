package data

import "github.com/emilaleksanteri/pubsub/internal/validator"

type Metadata struct {
	PageSize int `json:"page_size"`
}

func calculateMetadata(pageSize int) Metadata {
	if pageSize == 0 {
		return Metadata{}
	}

	return Metadata{
		PageSize: pageSize,
	}
}

type Filters struct {
	Take   int `json:"take"`
	Offset int `json:"offset"`
}

func ValidateFileters(v *validator.Validator, f Filters) {
	v.Check(f.Take > 0 && f.Take <= 20, "take", "must be between 1 and 20")
	v.Check(f.Offset >= 0, "offset", "must be greater than or equal to zero")
}
