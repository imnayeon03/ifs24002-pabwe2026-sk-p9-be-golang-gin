package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadDefaultDanOverride(t *testing.T) {
	t.Setenv("APP_PORT", "")
	assert.Equal(t, "8081", Load().AppPort)
	t.Setenv("APP_PORT", "9999")
	c := Load()
	assert.Equal(t, "9999", c.AppPort)
	assert.Contains(t, c.DSN(), "dbname=pabwe_p9_blog")
}
