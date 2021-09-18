module sigs.k8s.io/krew/site/functions

go 1.15

require (
	github.com/apex/gateway v1.1.2
	github.com/aws/aws-lambda-go v1.26.0 // indirect
	github.com/google/go-github/v39 v39.0.0
	github.com/pkg/errors v0.9.1
	golang.org/x/oauth2 v0.0.0-20180821212333-d2e6202438be
	golang.org/x/sync v0.0.0-20181221193216-37e7f081c4d4
	gopkg.in/yaml.v2 v2.4.0 // indirect
	sigs.k8s.io/krew v0.4.0
	sigs.k8s.io/yaml v1.2.0
)

// replace sigs.k8s.io/krew => ../../
