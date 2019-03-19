package ttgen

import (
	"gopkg.in/yaml.v2"
)

func LoadYAMLSitemap(data []byte) (Sitemap, error) {
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

func EncodeYAMLSitemap(s Sitemap) ([]byte, string, error) {
	d, err := yaml.Marshal(&s)
	return d, ".yaml", err
}
