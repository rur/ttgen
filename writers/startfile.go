package writers

import (
	"os"
	"path/filepath"
)

type startdata struct {
	Namespace string
	PageName  string
}

func WriteStartFile(dir string, pageName, namespace string) (string, error) {
	fileName := "start.go"
	filePath := filepath.Join(dir, "start.go")
	sf, err := os.Create(filePath)
	if err != nil {
		return fileName, err
	}
	defer sf.Close()

	err = startTemplate.Execute(sf, startdata{
		Namespace: namespace,
		PageName:  pageName,
	})
	return fileName, err
}
