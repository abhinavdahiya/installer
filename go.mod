module github.com/openshift/installer

require (
	github.com/Netflix/go-expect v0.0.0-20180928190340-9d1f4485533b // indirect
	github.com/ajeddeloh/go-json v0.0.0-20170920214419-6a2fe990e083 // indirect
	github.com/apparentlymart/go-cidr v1.0.0
	github.com/awalterschulze/gographviz v0.0.0-20170410065617-c84395e536e1
	github.com/aws/aws-sdk-go v1.15.67
	github.com/coreos/go-semver v0.2.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20181031085051-9002847aa142 // indirect
	github.com/coreos/ignition v0.26.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0 // indirect
	github.com/go-logr/zapr v0.1.0 // indirect
	github.com/gogo/protobuf v1.1.1 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b // indirect
	github.com/golang/groupcache v0.0.0-20181024230925-c65c006176ff // indirect
	github.com/google/btree v0.0.0-20180813153112-4030bb1f1f0c // indirect
	github.com/google/gofuzz v0.0.0-20170612174753-24818f796faf // indirect
	github.com/google/uuid v1.0.0 // indirect
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/gophercloud/gophercloud v0.0.0-20181114204705-3a7818a07cfc
	github.com/gophercloud/utils v0.0.0-20181029231510-34f5991525d1
	github.com/gregjones/httpcache v0.0.0-20180305231024-9cad4c3443a7
	github.com/hashicorp/golang-lru v0.5.0 // indirect
	github.com/hinshun/vt10x v0.0.0-20180809195222-d55458df857c // indirect
	github.com/imdario/mergo v0.3.6 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/json-iterator/go v1.1.5 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/kr/pty v1.1.3 // indirect
	github.com/libvirt/libvirt-go v4.8.0+incompatible
	github.com/mattn/go-colorable v0.0.9 // indirect
	github.com/mattn/go-isatty v0.0.4 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v0.0.0-20180701023420-4b7aa43c6742 // indirect
	github.com/onsi/ginkgo v1.7.0 // indirect
	github.com/onsi/gomega v1.4.3 // indirect
	github.com/openshift/cluster-api-provider-libvirt v0.0.0-20181101150541-437b1012ea0b
	github.com/openshift/cluster-network-operator v0.0.0-20181102160755-fb8b55a10724
	github.com/openshift/hive v0.0.0-20181101203307-8c7844d9b61c
	github.com/pborman/uuid v0.0.0-20180906182336-adf5a7427709
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/errors v0.8.0
	github.com/shurcooL/httpfs v0.0.0-20171119174359-809beceb2371 // indirect
	github.com/shurcooL/vfsgen v0.0.0-20181020040650-a97a25d856ca // indirect
	github.com/sirupsen/logrus v1.2.0
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3 // indirect
	github.com/stretchr/testify v1.2.2
	github.com/vincent-petithory/dataurl v0.0.0-20160330182126-9a301d65acbb
	go.uber.org/atomic v1.3.2 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.9.1 // indirect
	go4.org v0.0.0-20180809161055-417644f6feb5 // indirect
	golang.org/x/crypto v0.0.0-20181030102418-4d3f4d9ffa16
	golang.org/x/net v0.0.0-20181102091132-c10e9556a7bc // indirect
	golang.org/x/oauth2 v0.0.0-20181102170140-232e45548389 // indirect
	golang.org/x/sync v0.0.0-20181108010431-42b317875d0f // indirect
	golang.org/x/sys v0.0.0-20181031143558-9b800f95dbbc
	golang.org/x/time v0.0.0-20180412165947-fbb02b2291d2 // indirect
	google.golang.org/appengine v1.3.0 // indirect
	gopkg.in/AlecAivazis/survey.v1 v1.6.3
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/api v0.0.0-20181101211808-7c111c89a854
	k8s.io/apiextensions-apiserver v0.0.0-20181121072900-e8a638592964 // indirect
	k8s.io/apimachinery v0.0.0-20181022183627-f71dbbc36e12
	k8s.io/client-go v9.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20181114233023-0317810137be // indirect
	k8s.io/utils v0.0.0-20181022192358-4c3feeb576b0
	sigs.k8s.io/cluster-api v0.0.0-20181101193540-fee897706a82
	sigs.k8s.io/cluster-api-provider-aws v1.0.0-alpha.3.0.20181120224400-129f8fa1d402
	sigs.k8s.io/controller-runtime v0.1.7 // indirect
	sigs.k8s.io/testing_frameworks v0.0.0-20180709092217-5818a3a284a1 // indirect
)

replace sigs.k8s.io/cluster-api-provider-aws v1.0.0-alpha.3.0.20181120224400-129f8fa1d402 => github.com/openshift/cluster-api-provider-aws v0.1.1-0.20181115162746-e6986093d1fb
