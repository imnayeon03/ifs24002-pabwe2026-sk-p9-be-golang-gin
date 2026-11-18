package docs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSwaggerJSON(t *testing.T) {
	b, err := GetSwaggerJSON("localhost:9", "https")
	assert.NoError(t, err)
	assert.Contains(t, string(b), `"host":"localhost:9"`)
	assert.Contains(t, string(b), `"https"`)
}

func TestGetSwaggerJSONRusak(t *testing.T) {
	asli := specJSON
	defer func() { specJSON = asli }()
	specJSON = []byte("{")
	_, err := GetSwaggerJSON("h", "http")
	assert.Error(t, err)
}
