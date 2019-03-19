
## Treetop Generators

Quick and dirty application tools designed for the [Treetop](https://github.com/rur/treetop) project.
This was created to make it easier for me to experiment with different development patterns using the Treetop library.

### _Warning_

This is maintained as a personal tool, behavior may change without notice.

## CLI

### `ttsitemap` Command

Generate a functioning example site for a Treetop application given a potentially multi-page sitemap
definition. The files will be created in a temporary directory. The generated folder path is piped to stdout by default.

#### Example Usage

    # Usage: ttsitemap SITEMAP [OPTIONS]
    ttgen sitemap.yaml
    -> /tmp/12345678


#### Options:

`--human` Send human readable output to stdout

`--temp-dir DIR` Specify a directory to use as tmp for the purpose of generating files.

`--out-format FORMAT` Specify an out format for the routemap files. 'YAML' by default but 'TOML' is also supported.

### `ttroutes` Command

Generate a page routes.go file given a 'routemap' YAML file. This is the same format as the
the sitemap YAML but only the first page in the list will be adopted

#### Example Usage

    # Usage: ttroutes ROUTEMAP TEMPLATE DEST
    ttroutes routemap.yaml routes.go.templ routes.go
