package utils

import (
	"database/sql"
	"encoding/json"
)

func ToNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}

	return sql.NullString{String: s, Valid: s != ""}
}

func ToNullStringIfNotEmpty(s sql.NullString) sql.NullString {
	if !s.Valid || s.String == "" {
		return sql.NullString{Valid: false}
	}
	return s
}

func ToNullBoolFromSQL(b sql.NullBool) sql.NullBool {
	return b
}

// NullString is a wrapper for sql.NullString with JSON support
type NullString struct {
	sql.NullString
}

func (ns *NullString) UnmarshalJSON(data []byte) error {
	var s *string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s != nil && *s != "" {
		ns.Valid = true
		ns.String = *s
	} else {
		ns.Valid = false
	}
	return nil
}

// NullBool with JSON support
type NullBool struct {
	sql.NullBool
}

func (nb *NullBool) UnmarshalJSON(data []byte) error {
	var b *bool
	if err := json.Unmarshal(data, &b); err != nil {
		return err
	}
	if b != nil {
		nb.Valid = true
		nb.Bool = *b
	} else {
		nb.Valid = false
	}
	return nil
}

// NullFloat64 with JSON support
type NullFloat64 struct {
	sql.NullFloat64
}

func (nf *NullFloat64) UnmarshalJSON(data []byte) error {
	var f *float64
	if err := json.Unmarshal(data, &f); err != nil {
		return err
	}
	if f != nil {
		nf.Valid = true
		nf.Float64 = *f
	} else {
		nf.Valid = false
	}
	return nil
}
