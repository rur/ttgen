package ttgen

import (
	"bytes"

	"github.com/BurntSushi/toml"
)

func LoadTOMLSitemap(data []byte) (Sitemap, error) {
	var config Sitemap
	if err := toml.Unmarshal(data, &config); err != nil {
		return config, err
	}
	for i := 0; i < len(config.Pages); i++ {
		if config.Pages[i].Page == "" {
			config.Pages[i].Page = config.Pages[i].Name
		}
	}

	return config, nil
}

func EncodeTOMLSitemap(s Sitemap) (out []byte, ext string, err error) {
	buf := new(bytes.Buffer)
	err = toml.NewEncoder(buf).Encode(s)
	return buf.Bytes(), ".toml", err
}
