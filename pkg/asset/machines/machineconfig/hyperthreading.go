package machineconfig

import (
	"fmt"

	igntypes "github.com/coreos/ignition/config/v2_2/types"
	mcfgv1 "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/installer/pkg/asset/ignition"
	"github.com/openshift/installer/pkg/types"
)

// ForHyperthreading creates the MachineConfig to enable hyperthreading.
// RHCOS ships with pivot.service that uses the `/etc/pivot/kernel-args` to override the kernel arguments for hosts.
// TODO(abhinavdahiya): link to docs on `nosmt` kernel argument for RHEL 8
func ForHyperthreading(mode types.HyperthreadingMode, role string) *mcfgv1.MachineConfig {
	if mode == types.HyperthreadingEnabled {
		return nil
	}
	return &mcfgv1.MachineConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "machineconfiguration.openshift.io/v1",
			Kind:       "MachineConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("99-%s-hyperthreading", role),
			Labels: map[string]string{
				"machineconfiguration.openshift.io/role": role,
			},
		},
		Spec: mcfgv1.MachineConfigSpec{
			Config: igntypes.Config{
				Ignition: igntypes.Ignition{
					Version: igntypes.MaxVersion.String(),
				},
				Storage: igntypes.Storage{
					Files: []igntypes.File{
						ignition.FileFromString("/etc/pivot/kernel-args", "root", 0600, "nosmt"),
					},
				},
			},
		},
	}
}
