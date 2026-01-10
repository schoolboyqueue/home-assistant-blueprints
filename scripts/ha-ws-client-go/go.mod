module github.com/home-assistant-blueprints/ha-ws-client-go

go 1.25


require (
	github.com/gorilla/websocket v1.5.3
	github.com/home-assistant-blueprints/selfupdate v0.0.0
	github.com/home-assistant-blueprints/shared v0.0.0
	github.com/home-assistant-blueprints/testfixtures v0.0.0
	github.com/stretchr/testify v1.11.1
	github.com/urfave/cli/v3 v3.3.3
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/home-assistant-blueprints/testfixtures => ../testfixtures

replace github.com/home-assistant-blueprints/selfupdate => ../go-tools/selfupdate

replace github.com/home-assistant-blueprints/shared => ../internal/shared
