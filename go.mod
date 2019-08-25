module github.com/moolen/harbor-sync

go 1.12

require (
	github.com/go-logr/logr v0.1.0
	github.com/hashicorp/go-version v1.2.0
	github.com/mitchellh/hashstructure v1.0.0
	github.com/onsi/ginkgo v1.6.0
	github.com/onsi/gomega v1.4.2
	github.com/tomnomnom/linkheader v0.0.0-20180905144013-02ca5825eb80
	golang.org/x/sys v0.0.0-20190422165155-953cdadca894 // indirect
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.2.0
)
