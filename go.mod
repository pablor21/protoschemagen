module github.com/pablor21/protoschemagen

go 1.25.4

require (
	github.com/pablor21/gonnotation v0.0.2
	github.com/pablor21/goschemagen v0.0.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	golang.org/x/mod v0.30.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/tools v0.39.0 // indirect
)

replace github.com/pablor21/goschemagen => ../goschemagen

replace github.com/pablor21/gonnotation => ../gonnotation
