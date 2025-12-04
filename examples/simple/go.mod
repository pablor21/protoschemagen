module github.com/coso

go 1.25.4

require (
	google.golang.org/grpc v1.77.0
	google.golang.org/protobuf v1.36.10
)

require (
	github.com/pablor21/gonnotation v0.0.4 // indirect
	github.com/pablor21/goschemagen v0.0.6 // indirect
	github.com/pablor21/protoschemagen v0.0.2 // indirect
	golang.org/x/mod v0.30.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	golang.org/x/tools v0.39.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251022142026-3a174f9686a8 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/pablor21/protoschemagen => ../..

tool github.com/pablor21/protoschemagen
