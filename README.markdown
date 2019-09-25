
## Treetop Generators

Quick and dirty application tools designed for the [Treetop](https://github.com/rur/treetop) project.
This was created to make it easier for me to experiment with different development patterns using the Treetop library.

### _Warning_

This is maintained as a personal tool, behavior may change without notice.


## Command `ttpage`

Generate a functioning scaffold site based upon a routing configuration for a webpage.
The scaffold files will be created in a temporary directory and the temporary folder path is piped to stdout by default.

### Example Usage

    # Usage: ttpage SITEMAP [OPTIONS]
    ttpage routemap.toml
    -> /tmp/12345678


### Options:

`--human` Send human readable output to stdout

`--temp-dir DIR` Specify a directory to use as tmp for the purpose of generating files.

`--out-format FORMAT` Specify an out format for the routemap files. 'YAML' by default but 'TOML' is also supported.

## Command `ttroutes`

Generate an updated page routes.go file given a config file and a template.

### Example Usage

    # Usage: ttroutes ROUTEMAP TEMPLATE DEST
    ttroutes routemap.yaml routes.go.templ routes.go
