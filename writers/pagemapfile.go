package writers

import (
	"os"
	"path/filepath"
	text "text/template"

	generate "github.com/rur/ttgen"
)

func WritePagemapFiles(dir string, pageDef *generate.PartialDef, namespace string) ([]string, error) {
	var files []string
	pagemapName := "pagemap.yml"
	pagemapPath := filepath.Join(dir, "pagemap.yml")
	yf, err := os.Create(pagemapPath)
	if err != nil {
		return files, err
	}
	files = append(files, pagemapName)
	defer yf.Close()

	templateName := "routes.go.templ"
	templatePath := filepath.Join(dir, "routes.go.templ")
	tf, err := os.Create(templatePath)
	if err != nil {
		return files, err
	}
	files = append(files, templateName)
	defer tf.Close()

	if routesTemplateTemplate, err := text.New(templateName).Parse(routesTempl); err != nil {
		return files, err
	} else {
		t2 := routesTemplateTemplate.New("template")
		// tricky... replace template definition with a template definition!
		if _, err := t2.Delims("[[", "]]").Parse(`[[ block "routes" . ]]{{ block "routes" . }}{{ end }}[[ end ]]`); err != nil {
			return files, err
		} else if err = routesTemplateTemplate.Execute(tf, pageData{
			Namespace: namespace,
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
