package injector

import (
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/openservicemesh/osm/pkg/configurator"
)

func getInitContainerSpec(containerName string, cfg configurator.Configurator, outboundIPRangeExclusionList []string, outboundPortExclusionList []int,
	inboundPortExclusionList []int, enablePrivilegedInitContainer bool) corev1.Container {
	iptablesInitCommandsList := generateIptablesCommands(outboundIPRangeExclusionList, outboundPortExclusionList, inboundPortExclusionList)
	iptablesInitCommand := strings.Join(iptablesInitCommandsList, " && ")
	init := "if [ -z $(iptables -t nat -L | grep -c WSDONE) ]; then echo \"Iptables already inited. exiting\"; exit 0; fi" + " && " + iptablesInitCommand

	return corev1.Container{
		Name:  containerName,
		Image: cfg.GetInitContainerImage(),
		SecurityContext: &corev1.SecurityContext{
			Privileged: &enablePrivilegedInitContainer,
			Capabilities: &corev1.Capabilities{
				Add: []corev1.Capability{
					"NET_ADMIN",
				},
			},
		},
		Command: []string{"/bin/sh"},
		Args: []string{
			"-c",
			init,
		},
	}
}
