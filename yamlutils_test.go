package yamlutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	A string `json:"a"`
	B int    `json:"b"`
}

func TestLoadDocuments(t *testing.T) {
	assert := assert.New(t)

	yamlData := `
---
a: "foo"
b: 1

---
a: "bar"
b: 2
`

	var values []TestStruct
	err := LoadDocuments([]byte(yamlData), &values)
	if assert.NoError(err) {
		if assert.Len(values, 2) {
			assert.Equal("foo", values[0].A)
			assert.Equal(1, values[0].B)

			assert.Equal("bar", values[1].A)
			assert.Equal(2, values[1].B)
		}
	}

	var valuePointers []*TestStruct
	err = LoadDocuments([]byte(yamlData), &valuePointers)
	if assert.NoError(err) {
		if assert.Len(valuePointers, 2) {
			assert.Equal("foo", valuePointers[0].A)
			assert.Equal(1, valuePointers[0].B)

			assert.Equal("bar", valuePointers[1].A)
			assert.Equal(2, valuePointers[1].B)
		}
	}
}
