package ttgen

import (
	"gopkg.in/yaml.v2"
)

type Sitemap struct {
	Namespace string       `yaml:"namespace"`
	Pages     []PartialDef `yaml:"pages"`
}

type PartialDef struct {
	Page     string                   `yaml:"page,omitempty"`     // The name of the page
	Name     string                   `yaml:"name,omitempty"`     // The unique name for this view
	Fragment bool                     `yaml:"fragment,omitempty"` // Is this a 'FragmetOnly' route
	FullPage bool                     `yaml:"fullpage,omitempty"` // Is this a 'PageOnly' route
	Default  bool                     `yaml:"default,omitempty"`  // Is this a default subview
	Path     string                   `yaml:"path,omitempty"`     // determine if a HTTP route should be associated with this view
	Includes []string                 `yaml:"includes,omitempty"` // list of other partails that should be included in the route
	Handler  string                   `yaml:"handler,omitempty"`  // explicit handler declaration
	Template string                   `yaml:"template,omitempty"` // explicit template path
	Merge    string                   `yaml:"merge,omitempty"`    // treetop-merge attribute of a partial root element
	Method   string                   `yaml:"method,omitempty"`   // HTTP request method for a route, default "GET"
	Doc      string                   `yaml:"doc,omitempty"`      // Optional doc string to include with the generated handler
	Blocks   map[string][]*PartialDef `yaml:"blocks,omitempty"`   // List of subviews from this view
	URI      string                   `yaml:"uri,omitempty"`      // the entrypoint URL for the top level view
}

func LoadSitemap(data []byte) (Sitemap, error) {
	var config Sitemap
	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, err
	}
	for i := 0; i < len(config.Pages); i++ {
		if config.Pages[i].Page == "" {
			config.Pages[i].Page = config.Pages[i].Name
		}
	}

	return config, nil
}

func EncodeSitemap(s Sitemap) ([]byte, error) {
	d, err := yaml.Marshal(&s)
	return d, err
}
