package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"app/internal/config"
)

func TestMemoryMigrateSeed(t *testing.T) {
	db, err := ConnectMemory()
	require.NoError(t, err)
	require.NoError(t, Migrate(db))
	cfg := config.Config{AdminUsername: "admin", AdminPassword: "admin"}
	require.NoError(t, Seed(db, cfg))
	require.NoError(t, Seed(db, cfg)) // idempoten
	var n int64
	db.Model(&User{}).Count(&n)
	assert.Equal(t, int64(2), n)
}

func TestSeedError(t *testing.T) {
	db, _ := ConnectMemory() // belum dimigrasi → tabel tidak ada
	assert.Error(t, Seed(db, config.Config{AdminUsername: "a", AdminPassword: "b"}))
}

func TestSeedUserCreateError(t *testing.T) {
	db, _ := ConnectMemory()
	require.NoError(t, Migrate(db))
	assert.Error(t, seedUser(db, "x", string(make([]byte, 100)), "member")) // bcrypt >72 byte
}

func TestConnectPostgresGagal(t *testing.T) {
	_, err := Connect(config.Config{DBHost: "127.0.0.1", DBPort: "1", SSL: "disable"})
	assert.Error(t, err)
}
