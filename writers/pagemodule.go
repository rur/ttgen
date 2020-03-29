package writers

import (
	"os"
	"path/filepath"
)

func WriteMuxFile(dir string) (string, error) {
	fileName := "mux.go"
	filePath := filepath.Join(dir, "mux.go")
	sf, err := os.Create(filePath)
	if err != nil {
		return fileName, err
	}
	defer sf.Close()

	err = muxTemplate.Execute(sf, nil)
	if err != nil {
		return fileName, err
	}

	return fileName, nil
}
