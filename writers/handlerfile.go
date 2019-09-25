package writers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	generate "github.com/rur/ttgen"
)

type handlerBlockData struct {
	Identifier string
	Name       string
	FieldName  string
}

type handlerData struct {
	Info       string
	Type       string
	Doc        string
	Extends    string
	Blocks     []*handlerBlockData
	Identifier string
	Method     string
}

type handlersdata struct {
	Namespace     string
	PageName      string
	ViewHandlers  []handlerData
	BlockHandlers []handlerData
}

func WriteHandlerFile(dir string, views map[string]*generate.PartialDef, namespace, pageName string) (string, error) {
	fileName := "handlers.go"
	filePath := filepath.Join(dir, "handlers.go")
	data := handlersdata{
		Namespace: namespace,
		PageName:  pageName,
	}
	for _, view := range IterateSortedViews(views) {
		if viewHandler, blockHandlers, err := processViewHandlers(view, pageName); err != nil {
			return "", err
		} else {
			if viewHandler != nil {
				data.ViewHandlers = append(data.ViewHandlers, *viewHandler)
			}
			data.BlockHandlers = append(data.BlockHandlers, blockHandlers...)
		}
	}

	sf, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer sf.Close()

	err = handlerTemplate.Execute(sf, data)
	if err != nil {
		return fileName, err
	}

	return fileName, nil
}

func processViewHandlers(view *generate.PartialDef, pageName string) (*handlerData, []handlerData, error) {
	var viewHandler *handlerData
	var handlers []handlerData

	if view.Handler == "" {
		method := view.Method
		if method == "" {
			method = "GET"
		}
		viewName, err := SanitizeName(view.Name)
		if err != nil {
			return nil, nil, fmt.Errorf("Invalid view name '%s': %s", view.Name, err)
		}

		// base page handler
		viewHandler = &handlerData{
			Info:       viewName,
			Doc:        view.Doc,
			Type:       "(page)",
			Blocks:     make([]*handlerBlockData, 0, len(view.Blocks)),
			Identifier: viewName + "Handler",
			Method:     strings.ToUpper(method),
		}
	}

	blocks, err := IterateSortedBlocks(view.Blocks)
	if err != nil {
		return nil, nil, err
	}

	for _, block := range blocks {
		if viewHandler != nil {
			viewHandler.Blocks = append(viewHandler.Blocks, &handlerBlockData{
				Identifier: block.Ident + "Data",
				Name:       block.Name,
				FieldName:  generate.ValidPublicIdentifier(block.Name),
			})
		}

		for _, partial := range block.Partials {
			blockHandlers, err := processHandlersDef(block.Name, partial)
			if err != nil {
				return nil, nil, err
			}
			handlers = append(handlers, blockHandlers...)
		}
	}
	return viewHandler, handlers, nil
}

func processHandlersDef(blockName string, def *generate.PartialDef) ([]handlerData, error) {
	var handlers []handlerData
	var entryType string
	if def.Fragment {
		entryType = "(fragment)"
	} else if def.Default {
		entryType = "(default partial)"
	} else {
		entryType = "(partial)"
	}

	entryName, err := SanitizeName(def.Name)
	if err != nil {
		return handlers, fmt.Errorf("Invalid name '%s'", def.Name)
	}

	var handler *handlerData

	if def.Handler == "" {
		method := def.Method
		if method == "" {
			method = "GET"
		}
		// base page handler
		handler = &handlerData{
			Info:       entryName,
			Extends:    blockName,
			Doc:        def.Doc,
			Type:       entryType,
			Blocks:     make([]*handlerBlockData, 0, len(def.Blocks)),
			Identifier: entryName + "Handler",
			Method:     strings.ToUpper(method),
		}
	}

	blocks, err := IterateSortedBlocks(def.Blocks)
	if err != nil {
		return handlers, err
	}
	var nextHandlers []handlerData
	for _, block := range blocks {
		if handler != nil {
			handler.Blocks = append(handler.Blocks, &handlerBlockData{
				Identifier: block.Ident + "Data",
				Name:       block.Name,
				FieldName:  generate.ValidPublicIdentifier(block.Name),
			})
		}

		for _, partial := range block.Partials {
			blockHandlers, err := processHandlersDef(block.Ident, partial)
			if err != nil {
				return handlers, err
			}
			nextHandlers = append(nextHandlers, blockHandlers...)
		}
	}
	if handler != nil {
		handlers = append(handlers, *handler)
	}
	handlers = append(handlers, nextHandlers...)

	return handlers, nil
}
