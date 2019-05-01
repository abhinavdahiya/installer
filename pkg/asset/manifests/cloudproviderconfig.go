package manifests

import (
	"path/filepath"

	"github.com/ghodss/yaml"
	ospclientconfig "github.com/gophercloud/utils/openstack/clientconfig"
	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/installer/pkg/asset"
	"github.com/openshift/installer/pkg/asset/installconfig"
	icazure "github.com/openshift/installer/pkg/asset/installconfig/azure"
	osmachine "github.com/openshift/installer/pkg/asset/machines/openstack"
	"github.com/openshift/installer/pkg/asset/manifests/azure"
	vspheremanifests "github.com/openshift/installer/pkg/asset/manifests/vsphere"
	awstypes "github.com/openshift/installer/pkg/types/aws"
	azuretypes "github.com/openshift/installer/pkg/types/azure"
	libvirttypes "github.com/openshift/installer/pkg/types/libvirt"
	nonetypes "github.com/openshift/installer/pkg/types/none"
	openstacktypes "github.com/openshift/installer/pkg/types/openstack"
	vspheretypes "github.com/openshift/installer/pkg/types/vsphere"
)

var (
	cloudProviderConfigFileName = filepath.Join(manifestDir, "cloud-provider-config.yaml")
)

const (
	cloudProviderConfigDataKey = "config"
)

// CloudProviderConfig generates the cloud-provider-config.yaml files.
type CloudProviderConfig struct {
	ConfigMap *corev1.ConfigMap
	File      *asset.File
}

var _ asset.WritableAsset = (*CloudProviderConfig)(nil)

// Name returns a human friendly name for the asset.
func (*CloudProviderConfig) Name() string {
	return "Cloud Provider Config"
}

// Dependencies returns all of the dependencies directly needed to generate
// the asset.
func (*CloudProviderConfig) Dependencies() []asset.Asset {
	return []asset.Asset{
		&installconfig.InstallConfig{},
		&installconfig.ClusterID{},

		// PlatformCheck just checks the creds (and asks, if needed) and other platform configuration.
		// We do not actually use it in this asset directly, hence
		// it is put in the dependencies but not fetched in Generate
		&installconfig.PlatformCheck{},
	}
}

// Generate generates the CloudProviderConfig.
func (cpc *CloudProviderConfig) Generate(dependencies asset.Parents) error {
	installConfig := &installconfig.InstallConfig{}
	clusterID := &installconfig.ClusterID{}
	dependencies.Get(installConfig, clusterID)

	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "openshift-config",
			Name:      "cloud-provider-config",
		},
		Data: map[string]string{},
	}

	switch installConfig.Config.Platform.Name() {
	case awstypes.Name, libvirttypes.Name, nonetypes.Name:
		return nil
	case openstacktypes.Name:
		opts := &ospclientconfig.ClientOpts{}
		opts.Cloud = installConfig.Config.Platform.OpenStack.Cloud
		cloud, err := ospclientconfig.GetCloudFromYAML(opts)
		if err != nil {
			return errors.Wrap(err, "failed to get cloud config for openstack")
		}
		clouds := make(map[string]map[string]*ospclientconfig.Cloud)
		clouds["clouds"] = map[string]*ospclientconfig.Cloud{
			osmachine.CloudName: cloud,
		}
		marshalled, err := yaml.Marshal(clouds)
		if err != nil {
			return err
		}
		cm.Data[cloudProviderConfigDataKey] = string(marshalled)
	case azuretypes.Name:
		session, err := icazure.GetSession()
		if err != nil {
			return errors.Wrap(err, "could not get azure session")
		}
		azureConfig, err := azure.CloudProviderConfig{
			GroupLocation:  installConfig.Config.Azure.Region,
			ResourcePrefix: clusterID.InfraID,
			SubscriptionID: session.Credentials.SubscriptionID,
			TenantID:       session.Credentials.TenantID,
		}.JSON()
		if err != nil {
			return errors.Wrap(err, "could not create cloud provider config")
		}
		cm.Data[cloudProviderConfigDataKey] = azureConfig
	case vspheretypes.Name:
		vsphereConfig, err := vspheremanifests.CloudProviderConfig(
			installConfig.Config.ObjectMeta.Name,
			installConfig.Config.Platform.VSphere,
		)
		if err != nil {
			return errors.Wrap(err, "could not create cloud provider config")
		}
		cm.Data[cloudProviderConfigDataKey] = vsphereConfig
	default:
		return errors.New("invalid Platform")
	}

	cmData, err := yaml.Marshal(cm)
	if err != nil {
		return errors.Wrapf(err, "failed to create %s manifest", cpc.Name())
	}
	cpc.ConfigMap = cm
	cpc.File = &asset.File{
		Filename: cloudProviderConfigFileName,
		Data:     cmData,
	}
	return nil
}

// Files returns the files generated by the asset.
func (cpc *CloudProviderConfig) Files() []*asset.File {
	if cpc.File != nil {
		return []*asset.File{cpc.File}
	}
	return []*asset.File{}
}

// Load loads the already-rendered files back from disk.
func (cpc *CloudProviderConfig) Load(f asset.FileFetcher) (bool, error) {
	return false, nil
}
