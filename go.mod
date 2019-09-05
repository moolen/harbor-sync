module github.com/moolen/harbor-sync

go 1.12

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/go-logr/glogr v0.1.0
	github.com/go-logr/logr v0.1.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/mitchellh/hashstructure v1.0.0
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/prometheus/client_golang v0.9.0
	github.com/tomnomnom/linkheader v0.0.0-20180905144013-02ca5825eb80
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.2.0-beta.1.0.20190826154422-d90bbc6ec9fc
	sigs.k8s.io/controller-tools v0.2.0 // indirect
)
