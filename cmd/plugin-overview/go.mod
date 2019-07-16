module sigs.k8s.io/krew/cmd/plugin-overview

go 1.12

require (
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/pkg/errors v0.8.1 // indirect
	github.com/spf13/pflag v1.0.3
	sigs.k8s.io/krew v0.2.1
)

replace sigs.k8s.io/krew => ../..
