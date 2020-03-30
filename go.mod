module github.com/kudobuilder/test-tools

go 1.14

require (
	github.com/Masterminds/semver v1.5.0
	github.com/kudobuilder/kudo v0.11.1
	github.com/stretchr/testify v1.5.1
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v11.0.0+incompatible
)

replace (
	k8s.io/client-go => k8s.io/client-go v0.17.4
	k8s.io/kubectl => k8s.io/kubectl v0.17.4
)
