package main

import (
	"encoding/json"
	"os"
)

func writeJSON(data interface{}, filepath string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}
