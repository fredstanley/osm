// Package endpoint defines the interface for an endpoints provider. Endpoints providers communicate with the compute platforms
// and are primarily responsible for providing information regarding the endpoints for services, such as their IP
// addresses, port numbers and protocol information.
// Reference: https://github.com/openservicemesh/osm/blob/main/DESIGN.md#3-endpoints-providers
package endpoint

import (
	"fmt"
	"net"

	"github.com/openservicemesh/osm/pkg/identity"
	"github.com/openservicemesh/osm/pkg/service"
)

// Provider is an interface to be implemented by components abstracting Kubernetes, and other compute/cluster providers
type Provider interface {
	// ListEndpointsForService retrieves the IP addresses comprising the given service.
	ListEndpointsForService(service.MeshService) []Endpoint

	// ListEndpointsForIdentity retrieves the list of IP addresses for the given service account
	ListEndpointsForIdentity(identity.ServiceIdentity) []Endpoint

	// GetServicesForServiceAccount retrieves the namespaced services for a given service account
	GetServicesForServiceAccount(identity.K8sServiceAccount) ([]service.MeshService, error)

	// GetTargetPortToProtocolMappingForService returns a mapping of the service's ports to their corresponding application protocol
	GetTargetPortToProtocolMappingForService(service.MeshService) (map[uint32]string, error)

	// GetResolvableEndpointsForService returns the expected endpoints that are to be reached when the service FQDN is resolved under
	// the scope of the provider
	GetResolvableEndpointsForService(service.MeshService) ([]Endpoint, error)

	// GetID returns the unique identifier of the EndpointsProvider.
	GetID() string
}

//witesand change
// Endpoint is a tuple of IP and Port representing an instance of a service
type Endpoint struct {
	IP      net.IP `json:"ip"`
	Port    Port   `json:"port"`
	PodName string `json:"podName"`
}

func (ep Endpoint) String() string {
	//witesand
	return fmt.Sprintf("(ip=%s, port=%d, podName=%s)", ep.IP, ep.Port, ep.PodName)
}

// Port is a numerical type representing a port on which a service is exposed
type Port uint32
