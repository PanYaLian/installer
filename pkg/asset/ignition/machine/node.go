package machine

import (
	"fmt"
	"net/url"

	ignition "github.com/coreos/ignition/v2/config/v3_0/types"
	"github.com/vincent-petithory/dataurl"

	"github.com/openshift/installer/pkg/types"
	baremetaltypes "github.com/openshift/installer/pkg/types/baremetal"
	openstacktypes "github.com/openshift/installer/pkg/types/openstack"
	openstackdefaults "github.com/openshift/installer/pkg/types/openstack/defaults"
	ovirttypes "github.com/openshift/installer/pkg/types/ovirt"
	vspheretypes "github.com/openshift/installer/pkg/types/vsphere"
)

// pointerIgnitionConfig generates a config which references the remote config
// served by the machine config server.
func pointerIgnitionConfig(installConfig *types.InstallConfig, rootCA []byte, role string) *ignition.Config {
	var ignitionHost string
	// Default platform independent ignitionHost
	ignitionHost = fmt.Sprintf("api-int.%s:22623", installConfig.ClusterDomain())
	// Update ignitionHost as necessary for platform
	switch installConfig.Platform.Name() {
	case baremetaltypes.Name:
		// Baremetal needs to point directly at the VIP because we don't have a
		// way to configure DNS before Ignition runs.
		ignitionHost = fmt.Sprintf("%s:22623", installConfig.BareMetal.APIVIP)
	case openstacktypes.Name:
		apiVIP, err := openstackdefaults.APIVIP(installConfig.Networking)
		if err == nil {
			ignitionHost = fmt.Sprintf("%s:22623", apiVIP.String())
		}
	case ovirttypes.Name:
		ignitionHost = fmt.Sprintf("%s:22623", installConfig.Ovirt.APIVIP)
	case vspheretypes.Name:
		if installConfig.VSphere.APIVIP != "" {
			ignitionHost = fmt.Sprintf("%s:22623", installConfig.VSphere.APIVIP)
		}
	}

	mergeSourceURL := url.URL{
		Scheme: "https",
		Host:   ignitionHost,
		Path:   fmt.Sprintf("/config/%s", role),
	}
	mergeSource := mergeSourceURL.String()

	return &ignition.Config{
		Ignition: ignition.Ignition{
			Version: ignition.MaxVersion.String(),
			Config: ignition.IgnitionConfig{
				Merge: []ignition.ConfigReference{{
					Source: &mergeSource,
				}},
			},
			Security: ignition.Security{
				TLS: ignition.TLS{
					CertificateAuthorities: []ignition.CaReference{{
						Source: dataurl.EncodeBytes(rootCA),
					}},
				},
			},
		},
	}
}
