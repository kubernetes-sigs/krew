module sigs.k8s.io/krew/site/functions

go 1.15

require (
	github.com/apex/gateway v1.1.1
	github.com/aws/aws-lambda-go v1.19.1
	github.com/google/go-github/v32 v32.1.0
	github.com/pkg/errors v0.9.1
	golang.org/x/oauth2 v0.0.0-20180821212333-d2e6202438be
	golang.org/x/sync v0.0.0-20181221193216-37e7f081c4d4
	gopkg.in/yaml.v2 v2.2.8
	sigs.k8s.io/krew v0.4.0
	sigs.k8s.io/yaml v1.2.0
)

// replace sigs.k8s.io/krew => ../../
