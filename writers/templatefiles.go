package writers

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	generate "github.com/rur/ttgen"
)

type htmlBlockPartialData struct {
	Path     string
	Name     string
	Fragment bool
	Default  bool
	POSTOnly bool
}

type htmlBlockData struct {
	FieldName  string
	Identifier string
	Name       string
	Partials   []*htmlBlockPartialData
}

type partialData struct {
	Path     string
	Extends  string
	Fragment bool
	Name     string
	Merge    string
	Blocks   []*htmlBlockData
}

type indexSiteLinksData struct {
	URI    string
	Label  string
	Active bool
}

type indexData struct {
	Title     string
	SiteLinks []*indexSiteLinksData
	Blocks    []*htmlBlockData
}

func WriteIndexFile(dir string, view *generate.PartialDef, otherPages map[string]*generate.PartialDef) (string, error) {
	fileName := fmt.Sprintf("%s.html.tmpl", view.Name)
	filePath := filepath.Join(dir, fileName)
	sf, err := os.Create(filePath)
	if err != nil {
		return fileName, err
	}
	defer sf.Close()

	links := make([]*indexSiteLinksData, 0, len(otherPages))
	for _, other := range IterateSortedViews(otherPages) {
		if other.URI != "" {
			links = append(links, &indexSiteLinksData{
				URI:    other.URI,
				Label:  other.Name,
				Active: other.URI != view.URI,
			})
		}
	}

	blockList, err := IterateSortedBlocks(view.Blocks)
	if err != nil {
		return fileName, err
	}
	blocks := make([]*htmlBlockData, 0, len(blockList))
	for _, block := range blockList {
		blockData := htmlBlockData{
			FieldName:  generate.ValidPublicIdentifier(block.Name),
			Identifier: block.Ident,
			Name:       block.Name,
			Partials:   make([]*htmlBlockPartialData, 0, len(block.Partials)),
		}
		blocks = append(blocks, &blockData)
		for _, partial := range block.Partials {
			blockData.Partials = append(blockData.Partials, &htmlBlockPartialData{
				Path:     partial.Path,
				Name:     partial.Name,
				Fragment: partial.Fragment,
				Default:  partial.Default,
				POSTOnly: strings.ToUpper(partial.Method) == "POST",
			})
		}
	}

	err = indexTemplate.Execute(sf, indexData{
		Title:     view.Name,
		SiteLinks: links,
		Blocks:    blocks,
	})
	if err != nil {
		return fileName, err
	}

	return fileName, nil
}

func WriteTemplateBlock(dir string, blocks map[string][]*generate.PartialDef) ([]string, error) {
	var created []string
	blockList, err := IterateSortedBlocks(blocks)
	if err != nil {
		return created, err
	}
	for _, block := range blockList {
		blockTemplDir := path.Join(dir, block.Ident)
		if _, err := os.Stat(blockTemplDir); os.IsNotExist(err) {
			if err := os.Mkdir(blockTemplDir, os.ModePerm); err != nil {
				return created, fmt.Errorf("Error creating template dir '%s': %s", blockTemplDir, err)
			}
		}
		for _, def := range block.Partials {
			files, err := writePartialTemplate(blockTemplDir, def, block.Name)
			if err != nil {
				return created, err
			}
			for _, file := range files {
				created = append(created, path.Join(block.Ident, file))
			}
		}
	}
	return created, nil
}

func writePartialTemplate(dir string, def *generate.PartialDef, extends string) ([]string, error) {
	var created []string
	name, err := SanitizeName(def.Name)
	if err != nil {
		return created, fmt.Errorf("Invalid Partial name: '%s'", def.Name)
	}

	if def.Template == "" {
		// a template is not already defined, generate one
		partial := partialData{
			Path:     def.Path,
			Extends:  extends,
			Fragment: def.Fragment,
			Name:     def.Name,
			Merge:    def.Merge,
			Blocks:   make([]*htmlBlockData, 0, len(def.Blocks)),
		}
		blockList, err := IterateSortedBlocks(def.Blocks)
		if err != nil {
			return created, err
		}
		for _, block := range blockList {
			blockData := htmlBlockData{
				FieldName:  generate.ValidPublicIdentifier(block.Name),
				Identifier: block.Ident,
				Name:       block.Name,
				Partials:   make([]*htmlBlockPartialData, 0, len(block.Partials)),
			}
			for _, bPartial := range block.Partials {
				blockData.Partials = append(blockData.Partials, &htmlBlockPartialData{
					Path:     bPartial.Path,
					Name:     bPartial.Name,
					Fragment: bPartial.Fragment,
					Default:  bPartial.Default,
					POSTOnly: strings.ToUpper(bPartial.Method) == "POST",
				})
			}
			partial.Blocks = append(partial.Blocks, &blockData)
		}

		fileName := fmt.Sprintf("%s.html.tmpl", name)
		filePath := filepath.Join(dir, fileName)
		sf, err := os.Create(filePath)
		if err != nil {
			return created, err
		}
		defer sf.Close()
		created = append(created, fileName)

		err = partialTemplate.Execute(sf, partial)
		if err != nil {
			return created, fmt.Errorf("Error executing partial template '%s': %s", fileName, err)
		}
	}

	// writer nested templates
	files, err := WriteTemplateBlock(dir, def.Blocks)
	if err != nil {
		return created, nil
	}
	created = append(created, files...)

	return created, nil
}
