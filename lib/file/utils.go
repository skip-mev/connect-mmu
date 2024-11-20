package file

import (
	"encoding/json"
	"os"
)

func ReadJSONIntoFile[T any](filePath string) (T, error) {
	var t T
	bz, err := os.ReadFile(filePath)
	if err != nil {
		return t, err
	}
	err = json.Unmarshal(bz, &t)
	return t, err
}

func WriteJSONToFile(v any, path string) error {
	bz, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, bz, 0o600)
}
