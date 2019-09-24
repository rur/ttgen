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
Usage: ttpage page.toml [FLAGS...]
Create a temporary directory and generate templates and server code for given a site map.
By default the path to the new directory will be printed to stdout.

FLAGS:
--human	Human readable output
--temp-dir [DIRECTORY_PATH]	Path to directory that should be used as 'temp'
--out-format [FORMAT] output routemap in specified format, {yaml|toml}

`

func main() {
	if len(os.Args) < 2 {
		fmt.Printf(generateUsage)
		return
	}
	config := os.Args[1]
	var data []byte
	var err error
	if data, err = ioutil.ReadFile(config); err != nil {
		fmt.Printf("Error loading sitemap file: %v", err)
		return
	}
	var sitemap generate.RouteMap
	var decoder generate.RouteMapDecoder
	var inFormat string
	switch path.Ext(config) {
	case ".yml":
		decoder = generate.LoadYAMLRouteMap
		inFormat = "YAML"
	case ".yaml":
		decoder = generate.LoadYAMLRouteMap
		inFormat = "YAML"
	case ".tml":
		decoder = generate.LoadTOMLRouteMap
		inFormat = "TOML"
	case ".toml":
		decoder = generate.LoadTOMLRouteMap
		inFormat = "TOML"
	default:
		log.Fatalf("Unknown file extenstion for sitemap file %s", config)
	}
	if sitemap, err = decoder(data); err != nil {
		fmt.Printf("Error parsing sitemap %s: %v", inFormat, err)
		return
	}

	// cheap and cheerful arg parsing
	human := false
	skip := 0
	tmpDir := ""
	outFormat := "toml"
	for i, arg := range os.Args[2:] {
		if skip > 0 {
			skip = skip - 1
			continue
		} else if arg == "--human" {
			human = true
		} else if arg == "--temp-dir" {
			tmpDir = os.Args[i+3]
			skip = 1
		} else if arg == "--out-format" {
			outFormat = os.Args[i+3]
			skip = 1
		} else {
			log.Fatalf("Unknown flag '%s'\n\n%s", arg, generateUsage)
		}
	}
	var encoder generate.RouteMapEncoder
	switch strings.ToLower(outFormat) {
	case "yaml":
		encoder = generate.EncodeYAMLRouteMap
	case "toml":
		encoder = generate.EncodeTOMLRouteMap
	default:
		log.Fatalf("Invalid out format '%s', expecting YAML or TOML", outFormat)
	}

	outfolder, err := ioutil.TempDir(tmpDir, "")
	if err != nil {
		fmt.Printf("Error creating temp dir: %s", err)
		log.Fatalf("Treetop Page Generate FAILED")
	}

	createdFiles, err := generateAndWriteFiles(outfolder, sitemap, encoder)
	if err != nil {
		fmt.Printf("Treetop: Failed to build scaffold for sitemap %s\n Error: %s\n", config, err.Error())
		if err := os.RemoveAll(outfolder); err != nil {
			fmt.Printf("Scaffold failed but temp directory was not cleaned up: %s\n", err.Error())
		}
		log.Fatalf("Treetop Page Generate FAILED")
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

func generateAndWriteFiles(outDir string, page generate.RouteMap, encoder generate.RouteMapEncoder) ([]string, error) {
	var file string
	var err error
	created := make([]string, 0)

	// check that page namespace is a uri looking thing (without protocol, creds, etc...)
	// it will typically be something like "github.com/example/project"
	var nsReg = regexp.MustCompile(`(?i)^[A-Z][A-Z0-9-_]*(\.[A-Z][A-Z0-9-_]*)*(/[A-Z][A-Z0-9-_]*(\.[A-Z][A-Z0-9-_]*)*)*$`)
	if !nsReg.MatchString(page.Namespace) {
		return created, fmt.Errorf("Invalid site namespace in config: %s", page.Namespace)
	}

	appDir := filepath.Join(outDir, "app")
	if err := os.Mkdir(appDir, os.ModePerm); err != nil {
		return created, fmt.Errorf("Error creating 'app' dir in temp directory. %s", err)
	}

	pageDir := filepath.Join(outDir, "page")
	if err := os.Mkdir(pageDir, os.ModePerm); err != nil {
		return created, fmt.Errorf("Error creating 'page' dir in temp directory. %s", err)
	}

	pageName, err := writers.SanitizeName(page.Page)
	if err != nil {
		return created, err
	} else {
		// use sanitized name from here on
		page.Page = pageName
	}
	pagePkgDir := filepath.Join(pageDir, pageName)
	if err := os.Mkdir(pagePkgDir, os.ModePerm); err != nil {
		return created, fmt.Errorf("Error creating dir for page '%s'. %s", page.Page, err)
	}
	templatesDir := filepath.Join(pagePkgDir, "templates")
	if err := os.Mkdir(templatesDir, os.ModePerm); err != nil {
		return created, fmt.Errorf("Error creating template dir for page '%s'. %s", page.Page, err)
	}

	file, err = writers.WriteHandlerFile(pagePkgDir, page.Views, page.Namespace, pageName)
	if err != nil {
		return created, fmt.Errorf("Error creating handler.go file for page '%s'. %s", page.Page, err)
	}
	created = append(created, path.Join("page", pageName, file))

	for _, view := range writers.IterateSortedViews(page.Views) {
		if view.Template == "" {
			// only generate template file if sitemap doesn't have a template path already viewined
			file, err = writers.WriteIndexFile(templatesDir, view, page.Views)
			if err != nil {
				return created, fmt.Errorf("Error creating index template file for page '%s' view '%s'. %s", page.Page, view.Name, err)
			}
			created = append(created, path.Join("page", pageName, "templates", file))
		}

		tplPaths, err := writers.WriteTemplateBlock(templatesDir, view.Blocks)
		if err != nil {
			return created, fmt.Errorf("Error creating HTML partials for page '%s' view '%s'. %s", page.Page, view.Name, err)
		}
		for _, file = range tplPaths {
			created = append(created, path.Join("page", pageName, "templates", file))
		}
	}

	// NOTE: Routes need to done last due to mutation that happens to the sitemap definitions.
	// This is hack(y) code. The whole thing should be rewritten if the tool turns out to be useful in the future.
	file, err = writers.WriteRoutesFile(pagePkgDir, "routes.go", page.Views, page.Namespace, pageName, "")
	if err != nil {
		return created, fmt.Errorf("Error creating routes.go file for '%s'. %s", page.Page, err)
	}
	created = append(created, path.Join("page", pageName, file))

	rtmPaths, err := writers.WriteRoutemapFiles(pagePkgDir, page.Views, page.Namespace, pageName, encoder)
	if err != nil {
		return created, fmt.Errorf("Error creating routemap files for '%s'. %s", page.Page, err)
	}
	for _, file = range rtmPaths {
		created = append(created, path.Join("page", pageName, file))
	}

	file, err = writers.WriteContextFile(pageDir, page.Namespace)
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

	file, err = writers.WriteStartFile(outDir, page.Page, page.Namespace)
	if err != nil {
		return created, fmt.Errorf("Error creating start.go file. %s", err)
	}
	created = append(created, file)

	return created, nil
}
