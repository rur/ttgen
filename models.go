package ttgen

type RouteMap struct {
	Namespace string                 `yaml:"namespace" toml:"namespace"`
	Page      string                 `yaml:"page" toml:"page"`
	Views     map[string]*PartialDef `yaml:"views" toml:"views"`
}

type PartialDef struct {
	Name     string                   `yaml:"name,omitempty" toml:"name,omitempty"`         // The unique name for this view
	Fragment bool                     `yaml:"fragment,omitempty" toml:"fragment,omitempty"` // Is this a 'FragmetOnly' route
	FullPage bool                     `yaml:"fullpage,omitempty" toml:"fullpage,omitempty"` // Is this a 'PageOnly' route
	Default  bool                     `yaml:"default,omitempty" toml:"default,omitempty"`   // Is this a default subview
	Path     string                   `yaml:"path,omitempty" toml:"path,omitempty"`         // determine if a HTTP route should be associated with this view
	Includes []string                 `yaml:"includes,omitempty" toml:"includes,omitempty"` // list of other partails that should be included in the route
	Handler  string                   `yaml:"handler,omitempty" toml:"handler,omitempty"`   // explicit handler declaration
	Template string                   `yaml:"template,omitempty" toml:"template,omitempty"` // explicit template path
	Merge    string                   `yaml:"merge,omitempty" toml:"merge,omitempty"`       // treetop-merge attribute of a partial root element
	Method   string                   `yaml:"method,omitempty" toml:"method,omitempty"`     // HTTP request method for a route, default "GET"
	Doc      string                   `yaml:"doc,omitempty" toml:"doc,omitempty"`           // Optional doc string to include with the generated handler
	Blocks   map[string][]*PartialDef `yaml:"blocks,omitempty" toml:"blocks,omitempty"`     // List of subviews from this view
	URI      string                   `yaml:"uri,omitempty" toml:"uri,omitempty"`           // the entrypoint URL for the top level view
}

type RouteMapDecoder func([]byte) (RouteMap, error)
type RouteMapEncoder func(RouteMap) ([]byte, string, error)
