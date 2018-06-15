package main

import (
	"github.com/golang/glog"
	"github.com/google/krew/cmd/krew-manifest/plugin"
	"github.com/google/krew/cmd/krew/cmd"
)

func main() {
	cmd.Execute()
	defer glog.Flush()
}

func init() {
	cmd.InjectCommand(plugin.NewGenerateCmd())
}
