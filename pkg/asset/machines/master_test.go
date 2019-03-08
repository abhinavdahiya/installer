package machines

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"github.com/openshift/installer/pkg/asset"
	"github.com/openshift/installer/pkg/asset/ignition/machine"
	"github.com/openshift/installer/pkg/asset/installconfig"
	"github.com/openshift/installer/pkg/asset/rhcos"
	"github.com/openshift/installer/pkg/types"
	awstypes "github.com/openshift/installer/pkg/types/aws"
)

func TestMasterGenerate(t *testing.T) {
	cases := []struct {
		mode                  types.HyperthreadingMode
		expectedMachineConfig string
	}{
		{
			mode: types.HyperthreadingEnabled,
		},
		{
			mode: types.HyperthreadingDisabled,
			expectedMachineConfig: `---
apiVersion: machineconfiguration.openshift.io/v1
kind: MachineConfig
metadata:
  creationTimestamp: null
  labels:
    machineconfiguration.openshift.io/role: master
  name: 99-master-hyperthreading
spec:
  config:
    ignition:
      config: {}
      security:
        tls: {}
      timeouts: {}
      version: 2.2.0
    networkd: {}
    passwd: {}
    storage:
      files:
      - contents:
          source: data:text/plain;charset=utf-8;base64,
          verification: {}
        filesystem: root
        mode: 384
        path: /etc/default/rhcos/karg/nosmt
        user:
          name: root
    systemd: {}
  osImageURL: ""
`,
		},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("Hyperthreading %v", tc.mode), func(t *testing.T) {
			parents := asset.Parents{}
			parents.Add(
				&installconfig.ClusterID{
					UUID:    "test-uuid",
					InfraID: "test-infra-id",
				},
				&installconfig.InstallConfig{
					Config: &types.InstallConfig{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-cluster",
						},
						BaseDomain: "test-domain",
						Platform: types.Platform{
							AWS: &awstypes.Platform{
								Region: "us-east-1",
							},
						},
						ControlPlane: &types.MachinePool{
							Replicas:       pointer.Int64Ptr(1),
							Hyperthreading: tc.mode,
							Platform: types.MachinePoolPlatform{
								AWS: &awstypes.MachinePool{
									Zones: []string{"us-east-1a"},
								},
							},
						},
					},
				},
				(*rhcos.Image)(pointer.StringPtr("test-image")),
				&machine.Master{
					File: &asset.File{
						Filename: "master-ignition",
						Data:     []byte("test-ignition"),
					},
				},
			)
			master := &Master{}
			if err := master.Generate(parents); err != nil {
				t.Fatalf("failed to generate master machines: %v", err)
			}
			if tc.expectedMachineConfig != "" {
				if assert.Equal(t, 1, len(master.MachineConfigFiles), "expected one machine config file") {
					file := master.MachineConfigFiles[0]
					assert.Equal(t, "openshift/99_openshift-machineconfig_master.yaml", file.Filename, "unexpected machine config filename")
					assert.Equal(t, tc.expectedMachineConfig, string(file.Data), "unexepcted machine config contents")
				}
			} else {
				assert.Equal(t, 0, len(master.MachineConfigFiles), "expected no machine config files")
			}
		})
	}
}
