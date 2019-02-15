package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	generate "github.com/rur/ttgen"
	writers "github.com/rur/ttgen/writers"
)

var generateUsage = `
Usage: ttgen-routes pagemap.yml routes.go.templ ./routes.go
Write a routes file from pagemap yaml. This will overwrite an existing file at the destination


`

func main() {
	if len(os.Args) < 4 {
		fmt.Printf(generateUsage)
		return
	}
	yamlPath := os.Args[1]

	data, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		fmt.Printf("Error loading pagemap file: %v", err)
		return
	}
	pagemap, err := generate.LoadSitemap(data)
	if err != nil {
		fmt.Printf("Error parsing pagemap YAML: %v", err)
		return
	}

	templPath := os.Args[2]
	templData, err := ioutil.ReadFile(templPath)
	if err != nil {
		fmt.Printf("Error loading routes template file: %v", err)
		return
	}
	destPath := os.Args[3]

	createdFiles, err := generateAndWriteFiles(outfolder, sitemap)
	if err != nil {
		fmt.Printf("Treetop: Failed to build scaffold for sitemap %s\n Error: %s\n", config, err.Error())
		if err := os.RemoveAll(outfolder); err != nil {
			fmt.Printf("Scaffold failed but temp directory was not cleaned up: %s\n", err.Error())
		}
		return
	} else {
		// attempt to format the go code
		// this should not cause the generate command to fail if go fmt fails for some reason
		var fmtError []string
		for i := range createdFiles {
			if strings.HasSuffix(createdFiles[i], ".go") {
				cmd := exec.Command("go", "fmt", path.Join(outfolder, createdFiles[i]))
				output, err := cmd.CombinedOutput()
				if err != nil {
					fmtError = append(fmtError, fmt.Sprintf("%s Error: %s\nOutput: %s", createdFiles[i], err, string(output)))
				}
			}
		}
		if len(fmtError) > 0 {
			log.Fatalf(
				"Generated folder %s but `go fmt` failed for the following files:\n\t%s",
				outfolder,
				strings.Join(fmtError, "\n\t"),
			)
		}
	}

	if human {
		fmt.Printf("Generated Treetop file in folder: %s\n\nFiles:\n\t%s\n", outfolder, strings.Join(createdFiles, "\n\t"))
	} else {
		fmt.Print(outfolder)
	}
}

func generateAndWriteFiles(outDir string, sitemap generate.Sitemap) ([]string, error) {
	var file string
	var err error
	created := make([]string, 0)

	// check that sitemap namespace is a uri looking thing (without protocol, creds, etc...)
	// it will typically be something like "github.com/example/project"
	var nsReg = regexp.MustCompile(`(?i)^[A-Z][A-Z0-9-_]*(\.[A-Z][A-Z0-9-_]*)*(/[A-Z][A-Z0-9-_]*(\.[A-Z][A-Z0-9-_]*)*)*$`)
	if !nsReg.MatchString(sitemap.Namespace) {
		return created, fmt.Errorf("Invalid site namespace in config: %s", sitemap.Namespace)
	}

	appDir := filepath.Join(outDir, "app")
	if err := os.Mkdir(appDir, os.ModePerm); err != nil {
		return created, fmt.Errorf("Error creating 'app' dir in temp directory. %s", err)
	}

	pageDir := filepath.Join(outDir, "page")
	if err := os.Mkdir(pageDir, os.ModePerm); err != nil {
		return created, fmt.Errorf("Error creating 'page' dir in temp directory. %s", err)
	}

	for _, def := range sitemap.Pages {
		if def.Page == "" {
			def.Page = def.Name
		}
		pageName, err := writers.SanitizeName(def.Page)
		if err != nil {
			return created, err
		}
		pageDir := filepath.Join(pageDir, pageName)
		if err := os.Mkdir(pageDir, os.ModePerm); err != nil {
			return created, fmt.Errorf("Error creating dir for page '%s'. %s", def.Page, err)
		}
		templatesDir := filepath.Join(pageDir, "templates")
		if err := os.Mkdir(templatesDir, os.ModePerm); err != nil {
			return created, fmt.Errorf("Error creating template dir for page '%s'. %s", def.Page, err)
		}

		file, err = writers.WriteRoutesFile(pageDir, &def, sitemap.Namespace, pageName)
		if err != nil {
			return created, fmt.Errorf("Error creating routes.go file for '%s'. %s", def.Page, err)
		}
		created = append(created, path.Join("page", pageName, file))

		file, err = writers.WriteHandlerFile(pageDir, &def, sitemap.Namespace, pageName)
		if err != nil {
			return created, fmt.Errorf("Error creating handler.go file for page '%s'. %s", def.Page, err)
		}
		created = append(created, path.Join("page", pageName, file))

		if def.Template == "" {
			// only generate template file if sitemap doesn't have a template path already defined
			file, err = writers.WriteIndexFile(templatesDir, &def, sitemap.Pages)
			if err != nil {
				return created, fmt.Errorf("Error creating index.templ.html file for page '%s'. %s", def.Page, err)
			}
			created = append(created, path.Join("page", pageName, "templates", file))
		}

		files, err := writers.WriteTemplateBlock(templatesDir, def.Blocks)
		if err != nil {
			return created, fmt.Errorf("Error creating HTML partials for page '%s'. %s", def.Page, err)
		}
		for _, file = range files {
			created = append(created, path.Join("page", pageName, "templates", file))
		}
	}

	file, err = writers.WriteContextFile(pageDir, sitemap.Namespace)
	if err != nil {
		return created, fmt.Errorf("Error creating context.go file. %s", err)
	}
	created = append(created, path.Join("page", file))

	file, err = writers.WriteMuxFile(pageDir)
	if err != nil {
		return created, fmt.Errorf("Error creating mux.go file. %s", err)
	}
	created = append(created, path.Join("page", file))

	file, err = writers.WriteServerFile(appDir)
	if err != nil {
		return created, fmt.Errorf("Error creating server.go file. %s", err)
	}
	created = append(created, path.Join("app", file))

	file, err = writers.WriteResourcesFile(appDir)
	if err != nil {
		return created, fmt.Errorf("Error creating resources.go file. %s", err)
	}
	created = append(created, path.Join("app", file))

	file, err = writers.WriteStartFile(outDir, sitemap.Pages, sitemap.Namespace)
	if err != nil {
		return created, fmt.Errorf("Error creating start.go file. %s", err)
	}
	created = append(created, file)

	return created, nil
}