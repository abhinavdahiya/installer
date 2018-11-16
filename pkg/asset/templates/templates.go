// Package templates deals with creating template assets that will be used by other assets
package templates

import (
	"github.com/openshift/installer/pkg/asset"
	"github.com/openshift/installer/pkg/asset/templates/bootkube"
	"github.com/openshift/installer/pkg/asset/templates/tectonic"
)

// Templates is a collection of all the template assets
var Templates = []asset.Asset{
	&bootkube.KubeCloudConfig{},
	&bootkube.MachineConfigServerTLSSecret{},
	&bootkube.OpenshiftServiceCertSignerSecret{},
	&bootkube.Pull{},
	&bootkube.CVOOverrides{},
	&bootkube.LegacyCVOOverrides{},
	&bootkube.EtcdServiceEndpointsKubeSystem{},
	&bootkube.KubeSystemConfigmapEtcdServingCA{},
	&bootkube.KubeSystemConfigmapRootCA{},
	&bootkube.KubeSystemSecretEtcdClient{},
	&bootkube.OpenshiftWebConsoleNamespace{},
	&bootkube.OpenshiftMachineConfigOperator{},
	&bootkube.OpenshiftClusterAPINamespace{},
	&bootkube.OpenshiftServiceCertSignerNamespace{},
	&bootkube.EtcdServiceKubeSystem{},
	&tectonic.BindingDiscovery{},
	&tectonic.CloudCredsSecret{},
	&tectonic.RoleCloudCredsSecretReader{},
}
