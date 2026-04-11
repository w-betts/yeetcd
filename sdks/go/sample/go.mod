module github.com/yeetcd/yeetcd/sdk/sample

go 1.23

require (
	github.com/stretchr/testify v1.11.1
	github.com/yeetcd/yeetcd/sdk v0.0.0
	github.com/yeetcd/yeetcd/sdk/test v0.0.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/yeetcd/yeetcd/sdk => ../sdk

replace github.com/yeetcd/yeetcd/sdk/test => ../test
