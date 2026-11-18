// Package docs menyajikan spesifikasi OpenAPI. Berkas ini dibuat manual (swag init tidak dijalankan).
package docs

import (
	_ "embed"
	"encoding/json"
)

//go:embed swagger.json
var specJSON []byte

// GetSwaggerJSON mengembalikan spesifikasi dengan host & skema pemanggil.
func GetSwaggerJSON(host, scheme string) ([]byte, error) {
	var spec map[string]interface{}
	if err := json.Unmarshal(specJSON, &spec); err != nil {
		return nil, err
	}
	spec["host"] = host
	spec["schemes"] = []string{scheme}
	return json.Marshal(spec)
}
