package ttgen

import (
	"bytes"

	"github.com/BurntSushi/toml"
)

func LoadTOMLRouteMap(data []byte) (RouteMap, error) {
	var config RouteMap
	if err := toml.Unmarshal(data, &config); err != nil {
		return config, err
	}

	for name, def := range config.Views {
		def.Name = name
	}

	return config, nil
}

func EncodeTOMLRouteMap(s RouteMap) (out []byte, ext string, err error) {
	buf := new(bytes.Buffer)
	err = toml.NewEncoder(buf).Encode(s)
	return buf.Bytes(), ".toml", err
}
