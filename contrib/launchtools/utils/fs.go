package utils

import (
	"encoding/json"
	"os"
)

// WriteJSONConfig writes the given value as a pretty-printed JSON file.
func WriteJSONConfig(fileName string, v any) error {
	bz, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(fileName, bz, 0600)
}
