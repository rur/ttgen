package writers

import (
	"os"
	"path/filepath"

	generate "github.com/rur/ttgen"
)

type startdata struct {
	Namespace string
	Pages     []string
}

func WriteStartFile(dir string, pages []generate.PartialDef, namespace string) (string, error) {
	fileName := "start.go"
	filePath := filepath.Join(dir, "start.go")
	sf, err := os.Create(filePath)
	if err != nil {
		return fileName, err
	}
	defer sf.Close()

	pageNames := make([]string, len(pages))
	for i, def := range pages {
		pageNames[i] = def.Name
	}

	err = startTemplate.Execute(sf, startdata{
		Namespace: namespace,
		Pages:     pageNames,
	})
	if err != nil {
		return fileName, err
	}

	return fileName, nil
}
