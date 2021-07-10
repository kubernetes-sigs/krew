module sigs.k8s.io/krew

go 1.16

require (
	github.com/fatih/color v1.12.0
	github.com/google/go-cmp v0.5.6
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/mattn/go-isatty v0.0.13
	github.com/pkg/errors v0.9.1
	github.com/sahilm/fuzzy v0.1.0
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e // indirect
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/klog v1.0.0
	sigs.k8s.io/structured-merge-diff/v4 v4.1.2 // indirect
	sigs.k8s.io/yaml v1.2.0
)
