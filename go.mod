module github.com/apache/rocketmq-operator

go 1.16

require (
	github.com/google/uuid v1.1.2
	github.com/pkg/errors v0.9.1
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
	sigs.k8s.io/controller-runtime v0.10.0
)

replace github.com/apache/rocketmq-operator/ => ./
