package db

import "database/sql"

func NullStringFromStringPointer(s *string) (ns sql.NullString) {
	if s == nil {
		ns.Valid = false
		return
	}

	ns.String = *s
	ns.Valid = true
	return
}

func NullStringFromString(s string) (ns sql.NullString) {
	ns.String = s
	ns.Valid = true
	return
}

func PointerFromNullString(s sql.NullString) *string {
	if s.Valid {
		return &s.String
	}

	return nil
}

func PointerFromString(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}
