package ttgen

import (
	"gopkg.in/yaml.v2"
)

func LoadYAMLRouteMap(data []byte) (RouteMap, error) {
	var config RouteMap
	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, err
	}

	for name, def := range config.Views {
		def.Name = name
	}

	return config, nil
}

func EncodeYAMLRouteMap(s RouteMap) ([]byte, string, error) {
	d, err := yaml.Marshal(&s)
	return d, ".yaml", err
}
