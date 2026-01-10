module github.com/home-assistant-blueprints/validate-blueprint-go

go 1.25

require (
	github.com/fatih/color v1.18.0
	github.com/home-assistant-blueprints/selfupdate v0.0.0
	github.com/home-assistant-blueprints/shared v0.0.0
	github.com/home-assistant-blueprints/testfixtures v0.0.0
	github.com/stretchr/testify v1.11.1
	github.com/urfave/cli/v3 v3.3.3
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
)

replace github.com/home-assistant-blueprints/testfixtures => ../testfixtures

replace github.com/home-assistant-blueprints/selfupdate => ../go-tools/selfupdate

replace github.com/home-assistant-blueprints/shared => ../internal/shared
