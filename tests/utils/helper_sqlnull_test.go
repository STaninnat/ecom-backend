package utils_test

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/STaninnat/ecom-backend/utils"
	"github.com/stretchr/testify/assert"
)

func TestToNullString(t *testing.T) {
	assert.False(t, utils.ToNullString("").Valid)
	assert.True(t, utils.ToNullString("abc").Valid)
	assert.Equal(t, "abc", utils.ToNullString("abc").String)
}

func TestToNullStringIfNotEmpty(t *testing.T) {
	assert.False(t, utils.ToNullStringIfNotEmpty(sql.NullString{}).Valid)
	assert.False(t, utils.ToNullStringIfNotEmpty(sql.NullString{Valid: true, String: ""}).Valid)

	valid := sql.NullString{Valid: true, String: "ok"}
	result := utils.ToNullStringIfNotEmpty(valid)
	assert.True(t, result.Valid)
	assert.Equal(t, "ok", result.String)
}

func TestToNullBoolFromSQL(t *testing.T) {
	in := sql.NullBool{Valid: true, Bool: true}
	out := utils.ToNullBoolFromSQL(in)
	assert.Equal(t, in, out)
}

func TestNullString_UnmarshalJSON(t *testing.T) {
	var ns utils.NullString

	err := json.Unmarshal([]byte(`"hello"`), &ns)
	assert.NoError(t, err)
	assert.True(t, ns.Valid)
	assert.Equal(t, "hello", ns.String)

	err = json.Unmarshal([]byte(`null`), &ns)
	assert.NoError(t, err)
	assert.False(t, ns.Valid)

	err = json.Unmarshal([]byte(`""`), &ns)
	assert.NoError(t, err)
	assert.False(t, ns.Valid)
}

func TestNullBool_UnmarshalJSON(t *testing.T) {
	var nb utils.NullBool

	err := json.Unmarshal([]byte(`true`), &nb)
	assert.NoError(t, err)
	assert.True(t, nb.Valid)
	assert.True(t, nb.Bool)

	err = json.Unmarshal([]byte(`null`), &nb)
	assert.NoError(t, err)
	assert.False(t, nb.Valid)
}

func TestNullFloat64_UnmarshalJSON(t *testing.T) {
	var nf utils.NullFloat64

	err := json.Unmarshal([]byte(`123.45`), &nf)
	assert.NoError(t, err)
	assert.True(t, nf.Valid)
	assert.Equal(t, 123.45, nf.Float64)

	err = json.Unmarshal([]byte(`null`), &nf)
	assert.NoError(t, err)
	assert.False(t, nf.Valid)
}
