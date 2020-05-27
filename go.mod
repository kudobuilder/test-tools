module github.com/kudobuilder/test-tools

go 1.14

require (
	github.com/Masterminds/semver v1.5.0
	github.com/emicklei/go-restful v2.9.6+incompatible // indirect
	github.com/fatih/color v1.9.0 // indirect
	github.com/google/go-cmp v0.4.0 // indirect
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/kudobuilder/kudo v0.13.0
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/mattn/go-runewidth v0.0.4 // indirect
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v1.0.0 // indirect
	github.com/stretchr/testify v1.5.1
	golang.org/x/net v0.0.0-20200520182314-0ba52f642ac2 // indirect
	golang.org/x/sys v0.0.0-20200523222454-059865788121 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
	k8s.io/api v0.17.6
	k8s.io/apiextensions-apiserver v0.17.6 // indirect
	k8s.io/apimachinery v0.17.6
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/utils v0.0.0-20200520001619-278ece378a50 // indirect
)

replace (
	k8s.io/client-go => k8s.io/client-go v0.17.6
)
