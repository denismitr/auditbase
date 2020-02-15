package model

type ID string

func (id ID) Validate(v Validator) ValidationErrors {
	s := string(id)

	if v.IsEmptyString(s) {
		v.Add("ID", ":id cannot be empty")
	}

	if !v.IsUUID4(s) {
		v.Add("ID", ":id must be a valid UUID4 or be null for auto assigning")
	}

	return v.Errors()
}

func (id ID) String() string {
	return string(id)
}
