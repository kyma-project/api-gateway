package log

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func StructToPrettyJson(t *testing.T, v interface{}) string {
	t.Helper()
	str, err := json.MarshalIndent(v, "", "    ")
	assert.NoError(t, err)
	return string(str)
}
