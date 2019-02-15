package writers

import (
	"os"
	"path/filepath"
	text "text/template"

	generate "github.com/rur/ttgen"
)

const genGoTempl = `
package {{ .Pkg }}

//go:generate ttroutes ./routemap.yml ./routes.go.templ ./routes.go

`

func WriteRoutemapFiles(dir string, pageDef *generate.PartialDef, namespace, pageName string) ([]string, error) {
	var files []string
	routemapName := "routemap.yml"
	routemapPath := filepath.Join(dir, routemapName)
	yf, err := os.Create(routemapPath)
	if err != nil {
		return files, err
	}
	files = append(files, routemapName)
	defer yf.Close()

	templateName := "routes.go.templ"
	templatePath := filepath.Join(dir, "routes.go.templ")
	tf, err := os.Create(templatePath)
	if err != nil {
		return files, err
	}
	files = append(files, templateName)
	defer tf.Close()

	genName := "gen.go"
	genPath := filepath.Join(dir, "gen.go")
	genF, err := os.Create(genPath)
	if err != nil {
		return files, err
	}
	files = append(files, genName)
	defer genF.Close()

	// now use routes template file to generate another template
	if routesTemplateTemplate, err := text.New(templateName).Parse(routesTempl); err != nil {
		return files, err
	} else {
		t2 := routesTemplateTemplate.New("template")
		// replace template definition with another template definition.
		if _, err := t2.Delims("[[", "]]").Parse(`[[ block "routes" . ]]{{ block "routes" . }}{{ end }}[[ end ]]`); err != nil {
			return files, err
		} else if err = routesTemplateTemplate.Execute(tf, pageData{
			Namespace: namespace,
			Name:      pageName,
		}); err != nil {
			return files, err
		}
	}

	if enc, err := generate.EncodeSitemap(generate.Sitemap{
		Namespace: namespace,
		Pages:     []generate.PartialDef{*pageDef},
	}); err != nil {
		return files, err
	} else if _, err = yf.Write(enc); err != nil {
		return files, err
	}

	// now use routes template file to generate another template
	if genTemplate, err := text.New("gen.go").Parse(genGoTempl); err != nil {
		return files, err
	} else if err = genTemplate.Execute(genF, struct{ Pkg string }{
		Pkg: pageName,
	}); err != nil {
		return files, err
	}

	return files, nil
}
