module github.com/yeetcd/yeetcd/sdk/test

go 1.25.0

replace github.com/yeetcd/yeetcd/sdk => ../sdk

replace github.com/yeetcd/yeetcd => ../../../core

require (
	github.com/stretchr/testify v1.11.1
	github.com/yeetcd/yeetcd v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.79.2
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/net v0.51.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260209200024-4cfbd4190f57 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
