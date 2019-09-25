package writers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	textTemplate "text/template"

	generate "github.com/rur/ttgen"
)

type pageBlockData struct {
	Name string
}

type pageEntryData struct {
	Assignment string
	Name       string
	Extends    string
	Block      string
	Handler    string
	Type       string
	Template   string
	Path       string
}

type pageRouteData struct {
	Reference string
	Path      string
	Method    string
	Type      string
	Includes  []string
}

type pageTemplateData struct {
	Assignment string
	Path       string
	Type       string
}

type pageData struct {
	Namespace string
	Name      string
	Views     []pageEntryData
	Entries   []pageEntryData
	Routes    []pageRouteData
}

func WriteRoutesFile(dir string, fileName string, views map[string]*generate.PartialDef, namespace string, pageName string, overrideTempl string) (string, error) {
	page := pageData{
		Namespace: namespace,
		Name:      pageName,
	}

	for _, viewDef := range IterateSortedViews(views) {
		view, entries, routes, vEr := processViewDef(
			viewDef.Name,
			viewDef,
			filepath.Join("page", pageName, "templates"),
		)
		if vEr != nil {
			return "", vEr
		}
		page.Views = append(page.Views, view)
		page.Entries = append(page.Entries, entries...)
		page.Routes = append(page.Routes, routes...)
	}

	if len(page.Routes) == 0 {
		return fileName, fmt.Errorf("Page '%s' does not have any routes!", pageName)
	}

	// process includes in routes by scanning entries for matching paths
	pathMap := make(map[string]int)
	for index, en := range page.Entries {
		if en.Path != "" {
			pathMap[en.Path] = index
		}
	}
	for _, route := range page.Routes {
		for i, incl := range route.Includes {
			if j, ok := pathMap[incl]; ok {
				route.Includes[i] = page.Entries[j].Name
				page.Entries[j].Assignment = page.Entries[j].Name + " :="
			} else {
				return fileName, fmt.Errorf("Failed to match include path '%s' to a sub view entry for route '%s'", incl, route.Path)
			}
		}
	}

	filePath := filepath.Join(dir, "routes.go")
	sf, err := os.Create(filePath)
	if err != nil {
		return fileName, err
	}
	defer sf.Close()

	var tEr error
	if overrideTempl != "" {
		// user override for outer template of routes page
		var master *template.Template
		master, tEr = textTemplate.New("override").Parse(overrideTempl)
		if tEr != nil {
			return fileName, tEr
		}
		_, tEr = master.New("overlay").Parse(routesTempl)
		if tEr != nil {
			return fileName, tEr
		}
		tEr = master.ExecuteTemplate(sf, master.Name(), page)
	} else {
		tEr = routesTemplate.Execute(sf, page)
	}

	return fileName, tEr
}

func processViewDef(name string, def *generate.PartialDef, templatePath string) (pageEntryData, []pageEntryData, []pageRouteData, error) {
	var (
		view    pageEntryData
		entries []pageEntryData
		routes  []pageRouteData
		err     error
	)
	if view.Assignment, err = SanitizeName(name); err != nil {
		return view, entries, routes, err
	} else {
		view.Assignment = view.Assignment + "View"
	}

	sortedBlocks, err := IterateSortedBlocks(def.Blocks)
	if err != nil {
		return view, entries, routes, err
	}
	for _, block := range sortedBlocks {

		entries = append(entries, pageEntryData{
			Name: block.Name,
			Type: "Spacer",
		})

		for i, partial := range block.Partials {
			blockEntries, blockRoutes, err := processEntries(
				view.Assignment,
				block.Name,
				[]string{def.Name, partial.Name},
				partial,
				filepath.Join(templatePath, block.Ident),
				block.Name,
			)
			if err != nil {
				return view, entries, routes, err
			}
			entries = append(entries, blockEntries...)
			routes = append(routes, blockRoutes...)
			if len(blockEntries) > 1 && i < len(block.Partials)-1 {
				entries = append(entries, pageEntryData{
					Name: block.Name,
					Type: "Spacer",
				})
			}
		}
	}

	if def.Path != "" {
		route := pageRouteData{
			Reference: view.Assignment,
			Path:      strings.Trim(def.Path, " "),
			Type:      "Page",
			Includes:  append([]string{}, def.Includes...),
		}

		if def.Method == "" {
			route.Method = "GET"
		} else if def.Method == "any" {
			route.Method = ""
		} else {
			route.Method = strings.ToUpper(def.Method)
		}
		routes = append(routes, route)
	}

	view.Handler = def.Handler
	if view.Handler == "" {
		view.Handler = fmt.Sprintf("cxt.Bind(%sHandler)", def.Name)
	}

	view.Template = def.Template
	if view.Template == "" {
		view.Template = filepath.Join(templatePath, fmt.Sprintf("%s.html.tmpl", def.Name))
	}
	// also update definition
	def.Handler = view.Handler
	def.Template = view.Template

	return view, entries, routes, err
}

func processEntries(extends, blockName string, names []string, def *generate.PartialDef, templatePath string, seen ...string) ([]pageEntryData, []pageRouteData, error) {
	var entryType string
	var entries []pageEntryData
	var routes []pageRouteData

	if def.Default {
		entryType = "DefaultSubView"
	} else {
		entryType = "SubView"
	}

	entryName, err := SanitizeName(def.Name)
	if err != nil {
		return entries, routes, fmt.Errorf("Invalid %s name '%s' @ %s", entryType, def.Name, strings.Join(seen, " -> "))
	}

	handler := def.Handler
	if handler == "" {
		handler = fmt.Sprintf("cxt.Bind(%sHandler)", entryName)
	}

	template := def.Template
	if template == "" {
		template = filepath.Join(templatePath, entryName+".html.tmpl")
	}

	entry := pageEntryData{
		Name:     entryName,
		Extends:  extends,
		Block:    blockName,
		Handler:  handler,
		Type:     entryType,
		Template: template,
		Path:     strings.Join(names, " > "),
	}

	def.Template = template
	def.Handler = handler

	// the assignment for an entry must be blanked if there are no routes or subviews
	// assignment may be reinstated if this entry is used as an include for another route
	if def.Path == "" && len(def.Blocks) == 0 {
		entry.Assignment = "_ ="
	} else {
		entry.Assignment = entryName + " :="
	}

	entries = append(entries, entry)

	sortedBlocks, err := IterateSortedBlocks(def.Blocks)
	if err != nil {
		return entries, routes, err
	}
	for _, block := range sortedBlocks {
		if len(block.Partials) == 0 {
			continue
		}
		entries = append(entries, pageEntryData{
			Name: strings.Join(append(seen, block.Name), " -> "),
			Type: "Spacer",
		})

		for _, partial := range block.Partials {
			blockEntries, blockRoutes, err := processEntries(
				entry.Name,
				block.Name,
				append(names, partial.Name),
				partial,
				filepath.Join(templatePath, block.Ident),
				append(seen, block.Name)...,
			)
			if err != nil {
				return entries, routes, err
			}
			entries = append(entries, blockEntries...)
			routes = append(routes, blockRoutes...)
		}
	}

	if def.Path != "" {
		route := pageRouteData{
			Reference: entryName,
			Path:      strings.Trim(def.Path, " "),
			Includes:  append([]string{}, def.Includes...),
		}
		if def.Fragment {
			route.Type = "Fragment"
		} else if def.FullPage {
			route.Type = "Page"
		} else {
			route.Type = "Partial"
		}
		if def.Method == "" {
			route.Method = "GET"
		} else if def.Method == "any" {
			route.Method = ""
		} else {
			route.Method = strings.ToUpper(def.Method)
		}
		routes = append(routes, route)
	}

	return entries, routes, nil
}
