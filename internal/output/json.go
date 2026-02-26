package output

import (
	"encoding/json"
	"os"

	"k8s-recovery-visualizer/internal/model"
)

func WriteJSON(path string, b *model.Bundle) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(b)
}
