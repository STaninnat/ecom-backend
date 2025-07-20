package utils

import (
	"database/sql"
	"encoding/json"
)

// ToNullString returns a sql.NullString that is valid if s is not empty.
func ToNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

// ToNullStringIfNotEmpty returns a NullString only if it is valid and not empty.
func ToNullStringIfNotEmpty(s sql.NullString) sql.NullString {
	s.Valid = s.Valid && s.String != ""
	return s
}

// ToNullBoolFromSQL returns the given sql.NullBool as is.
func ToNullBoolFromSQL(b sql.NullBool) sql.NullBool {
	return b
}

// NullString is a wrapper for sql.NullString with JSON marshaling and unmarshaling support.
type NullString struct {
	sql.NullString
}

// MarshalJSON implements json.Marshaler for NullString.
func (ns NullString) MarshalJSON() ([]byte, error) {
	if ns.Valid {
		return json.Marshal(ns.String)
	}
	return json.Marshal(nil)
}

// NullBool is a wrapper for sql.NullBool with JSON marshaling and unmarshaling support.
type NullBool struct {
	sql.NullBool
}

// MarshalJSON implements json.Marshaler for NullBool.
func (nb NullBool) MarshalJSON() ([]byte, error) {
	if nb.Valid {
		return json.Marshal(nb.Bool)
	}
	return json.Marshal(nil)
}

// NullFloat64 is a wrapper for sql.NullFloat64 with JSON marshaling and unmarshaling support.
type NullFloat64 struct {
	sql.NullFloat64
}

// MarshalJSON implements json.Marshaler for NullFloat64.
func (nf NullFloat64) MarshalJSON() ([]byte, error) {
	if nf.Valid {
		return json.Marshal(nf.Float64)
	}
	return json.Marshal(nil)
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
