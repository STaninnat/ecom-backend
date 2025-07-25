// Package utils provides utility functions and helpers used throughout the ecom-backend project.
package utils

import (
	"database/sql"
	"encoding/json"
	"testing"
)

// null_types_test.go: Tests for nullable types and SQL/JSON marshaling helpers.

// TestToNullString tests the ToNullString function for converting strings to sql.NullString.
func TestToNullString(t *testing.T) {
	t.Helper()
	cases := []struct {
		in   string
		want sql.NullString
	}{
		{"", sql.NullString{String: "", Valid: false}},
		{"foo", sql.NullString{String: "foo", Valid: true}},
	}
	for _, c := range cases {
		got := ToNullString(c.in)
		if got != c.want {
			t.Errorf("ToNullString(%q) = %+v, want %+v", c.in, got, c.want)
		}
	}
}

// TestToNullStringIfNotEmpty tests the ToNullStringIfNotEmpty function for handling empty and non-empty sql.NullString values.
func TestToNullStringIfNotEmpty(t *testing.T) {
	t.Helper()
	cases := []struct {
		in   sql.NullString
		want sql.NullString
	}{
		{sql.NullString{String: "", Valid: false}, sql.NullString{String: "", Valid: false}},
		{sql.NullString{String: "foo", Valid: true}, sql.NullString{String: "foo", Valid: true}},
		{sql.NullString{String: "", Valid: true}, sql.NullString{String: "", Valid: false}},
		{sql.NullString{String: "bar", Valid: false}, sql.NullString{String: "bar", Valid: false}},
	}
	for _, c := range cases {
		got := ToNullStringIfNotEmpty(c.in)
		if got != c.want {
			t.Errorf("ToNullStringIfNotEmpty(%+v) = %+v, want %+v", c.in, got, c.want)
		}
	}
}

// TestToNullBoolFromSQL tests the ToNullBoolFromSQL function for correct passthrough of sql.NullBool values.
func TestToNullBoolFromSQL(t *testing.T) {
	t.Helper()
	b := sql.NullBool{Bool: true, Valid: true}
	if got := ToNullBoolFromSQL(b); got != b {
		t.Errorf("ToNullBoolFromSQL(%+v) = %+v, want %+v", b, got, b)
	}
}

// TestNullStringJSON tests the NullString type's JSON marshaling and unmarshaling behavior.
func TestNullStringJSON(t *testing.T) {
	t.Helper()
	cases := []struct {
		name string
		val  NullString
		want string
	}{
		{"valid", NullString{sql.NullString{String: "foo", Valid: true}}, `"foo"`},
		{"invalid", NullString{sql.NullString{String: "", Valid: false}}, `null`},
	}
	for _, c := range cases {
		b, err := json.Marshal(c.val)
		if err != nil {
			t.Errorf("Marshal error: %v", err)
		}
		if string(b) != c.want {
			t.Errorf("Marshal got %q, want %q", string(b), c.want)
		}
	}

	// Unmarshal valid string
	var ns NullString
	err := json.Unmarshal([]byte(`"bar"`), &ns)
	if err != nil || !ns.Valid || ns.String != "bar" {
		t.Errorf("Unmarshal valid string failed: %+v, err=%v", ns, err)
	}
	// Unmarshal null
	err = json.Unmarshal([]byte(`null`), &ns)
	if err != nil || ns.Valid {
		t.Errorf("Unmarshal null failed: %+v, err=%v", ns, err)
	}
	// Unmarshal empty string
	err = json.Unmarshal([]byte(`""`), &ns)
	if err != nil || ns.Valid {
		t.Errorf("Unmarshal empty string failed: %+v, err=%v", ns, err)
	}
	// Unmarshal invalid type
	err = json.Unmarshal([]byte(`123`), &ns)
	if err == nil {
		t.Errorf("expected error for invalid type, got nil")
	}
}

// TestNullBoolJSON tests the NullBool type's JSON marshaling and unmarshaling behavior.
func TestNullBoolJSON(t *testing.T) {
	t.Helper()
	cases := []struct {
		name string
		val  NullBool
		want string
	}{
		{"valid", NullBool{sql.NullBool{Bool: true, Valid: true}}, `true`},
		{"invalid", NullBool{sql.NullBool{Bool: false, Valid: false}}, `null`},
	}
	for _, c := range cases {
		b, err := json.Marshal(c.val)
		if err != nil {
			t.Errorf("Marshal error: %v", err)
		}
		if string(b) != c.want {
			t.Errorf("Marshal got %q, want %q", string(b), c.want)
		}
	}

	// Unmarshal valid bool
	var nb NullBool
	err := json.Unmarshal([]byte(`true`), &nb)
	if err != nil || !nb.Valid || nb.Bool != true {
		t.Errorf("Unmarshal valid bool failed: %+v, err=%v", nb, err)
	}
	// Unmarshal null
	err = json.Unmarshal([]byte(`null`), &nb)
	if err != nil || nb.Valid {
		t.Errorf("Unmarshal null failed: %+v, err=%v", nb, err)
	}
	// Unmarshal invalid type
	err = json.Unmarshal([]byte(`123`), &nb)
	if err == nil {
		t.Errorf("expected error for invalid type, got nil")
	}
}

// TestNullFloat64JSON tests the NullFloat64 type's JSON marshaling and unmarshaling behavior.
func TestNullFloat64JSON(t *testing.T) {
	t.Helper()
	cases := []struct {
		name string
		val  NullFloat64
		want string
	}{
		{"valid", NullFloat64{sql.NullFloat64{Float64: 1.23, Valid: true}}, `1.23`},
		{"invalid", NullFloat64{sql.NullFloat64{Float64: 0, Valid: false}}, `null`},
	}
	for _, c := range cases {
		b, err := json.Marshal(c.val)
		if err != nil {
			t.Errorf("Marshal error: %v", err)
		}
		if string(b) != c.want {
			t.Errorf("Marshal got %q, want %q", string(b), c.want)
		}
	}

	// Unmarshal valid float
	var nf NullFloat64
	err := json.Unmarshal([]byte(`1.23`), &nf)
	if err != nil || !nf.Valid || nf.Float64 != 1.23 {
		t.Errorf("Unmarshal valid float failed: %+v, err=%v", nf, err)
	}
	// Unmarshal null
	err = json.Unmarshal([]byte(`null`), &nf)
	if err != nil || nf.Valid {
		t.Errorf("Unmarshal null failed: %+v, err=%v", nf, err)
	}
	// Unmarshal invalid type
	err = json.Unmarshal([]byte(`"foo"`), &nf)
	if err == nil {
		t.Errorf("expected error for invalid type, got nil")
	}
}
