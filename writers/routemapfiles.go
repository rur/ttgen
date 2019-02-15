package writers

import (
	"os"
	"path/filepath"
	text "text/template"

	generate "github.com/rur/ttgen"
)

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
	return files, nil
}
