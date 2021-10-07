package lds

import (
	xds_discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/openservicemesh/osm/pkg/identity"

	"github.com/openservicemesh/osm/pkg/catalog"
	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/configurator"
	"github.com/openservicemesh/osm/pkg/envoy"
	"github.com/openservicemesh/osm/pkg/envoy/registry"
)

const (
	inboundListenerName    = "inbound_listener"
	outboundListenerName   = "outbound_listener"
	prometheusListenerName = "inbound_prometheus_listener"
)

// NewResponse creates a new Listener Discovery Response.
// The response build 3 Listeners:
// 1. Inbound listener to handle incoming traffic
// 2. Outbound listener to handle outgoing traffic
// 3. Prometheus listener for metrics
func NewResponse(meshCatalog catalog.MeshCataloger, proxy *envoy.Proxy, _ *xds_discovery.DiscoveryRequest, cfg configurator.Configurator, _ certificate.Manager, proxyRegistry *registry.ProxyRegistry) ([]types.Resource, error) {
//func NewResponse(meshCatalog catalog.MeshCataloger, proxy *envoy.Proxy, _ *xds_discovery.DiscoveryRequest, cfg configurator.Configurator, _ certificate.Manager) (*xds_discovery.DiscoveryResponse, error) {
	svcList, err := proxyRegistry.ListProxyServices(proxy)
	//svcList, err := meshCatalog.GetServicesFromEnvoyCertificate(proxy.GetCommonName())
	if err != nil {
		log.Error().Err(err).Msgf("Error looking up MeshService for Envoy with CN=%q", proxy.GetCertificateCommonName())
		return nil, err
	}
	// Github Issue #1575
	proxyServiceName := svcList[0]

	key := proxy.GetPodUID()+ "lds"
	r, found := meshCatalog.GetWitesandCache(key)
	if found {
		return r, nil
	}

	svcAccount, err := envoy.GetServiceAccountFromProxyCertificate(proxy.GetCertificateCommonName())
//	svcAccount, err := catalog.GetServiceAccountFromProxyCertificate(proxy.GetCommonName())
	if err != nil {
		log.Error().Err(err).Msgf("Error retrieving SerivceAccount for proxy %s", proxy.GetCertificateCommonName())
		return nil, err
	}

	var ldsResources []types.Resource

	//resp := &xds_discovery.DiscoveryResponse{
	//	TypeUrl: string(envoy.TypeLDS),
	//}

	var statsHeaders map[string]string
	if featureflags := cfg.GetFeatureFlags(); featureflags.EnableWASMStats {
		statsHeaders = proxy.StatsHeaders()
	}
	//lb := newListenerBuilder(meshCatalog, svcAccount, cfg)
	lb := newListenerBuilder(meshCatalog, svcAccount.ToServiceIdentity(), cfg, statsHeaders)


	// --- OUTBOUND -------------------
	outboundListener, err := lb.newOutboundListener(svcList)
	if err != nil {
		log.Error().Err(err).Msgf("Error making outbound listener config for proxy %s", proxyServiceName)
	} else {
		if outboundListener == nil {
			log.Debug().Msgf("Not programming Outbound listener for proxy %s", proxyServiceName)
		} else {
			ldsResources = append(ldsResources, outboundListener)

			//if marshalledOutbound, err := ptypes.MarshalAny(outboundListener); err != nil {
			//	log.Error().Err(err).Msgf("Failed to marshal outbound listener config for proxy %s", proxyServiceName)
			//} else {
			//	resp.Resources = append(resp.Resources, marshalledOutbound)
			//}
		}
	}

	// --- INBOUND -------------------
	inboundListener := newInboundListener()
	// --- INBOUND: mesh filter chain
	inboundMeshFilterChains, err := lb.getInboundInMeshFilterChain(proxyServiceName)
	if  err == nil {
		inboundListener.FilterChains = append(inboundListener.FilterChains, inboundMeshFilterChains)
	} else {
		log.Error().Err(err).Msgf("Error making inbound listener config for proxy %s", proxyServiceName)
	}

	// --- INGRESS -------------------
	// Apply an ingress filter chain if there are any ingress routes
        /* Ingress rules are taken care by iptables, having them here
           causes duplicates without TLS configuration.

	if ingressRoutesPerHost, err := meshCatalog.GetIngressRoutesPerHost(proxyServiceName); err != nil {
		log.Error().Err(err).Msgf("Error getting ingress routes per host for service %s", proxyServiceName)
	} else {
		thereAreIngressRoutes := len(ingressRoutesPerHost) > 0

		if thereAreIngressRoutes {
			log.Info().Msgf("Found k8s Ingress for MeshService %s, applying necessary filters", proxyServiceName)
			// This proxy is fronting a service that is a backend for an ingress, add a FilterChain for it
			ingressFilterChains := lb.getIngressFilterChains(proxyServiceName)
			inboundListener.FilterChains = append(inboundListener.FilterChains, ingressFilterChains...)
		} else {
			log.Trace().Msgf("There is no k8s Ingress for service %s", proxyServiceName)
		}
	}
	*/

	if len(inboundListener.FilterChains) > 0 {
		ldsResources = append(ldsResources, inboundListener)

		// Inbound filter chains can be empty if the there both ingress and in-mesh policies are not configued.
		// Configuring a listener without a filter chain is an error.
		//if marshalledInbound, err := ptypes.MarshalAny(inboundListener); err != nil {
		//	log.Error().Err(err).Msgf("Error marshalling inbound listener config for proxy %s", proxyServiceName)
		//} else {
		//	resp.Resources = append(resp.Resources, marshalledInbound)
		//}
	}

	//if cfg.IsPrometheusScrapingEnabled() {
	//	// Build Prometheus listener config
	//	prometheusConnManager := getPrometheusConnectionManager(prometheusListenerName, constants.PrometheusScrapePath, constants.EnvoyMetricsCluster)
	//	if prometheusListener, err := buildPrometheusListener(prometheusConnManager); err != nil {
	//		log.Error().Err(err).Msgf("Error building Prometheus listener config for proxy %s", proxyServiceName)
	//	} else {
	//		if marshalledPrometheus, err := ptypes.MarshalAny(prometheusListener); err != nil {
	//			log.Error().Err(err).Msgf("Error marshalling Prometheus listener config for proxy %s", proxyServiceName)
	//		} else {
	//			resp.Resources = append(resp.Resources, marshalledPrometheus)
	//		}
	//	}
	//}

	meshCatalog.SetWitesandCache(key, ldsResources)
	return ldsResources, nil
}

//func newListenerBuilder(meshCatalog catalog.MeshCataloger, svcAccount service.K8sServiceAccount, cfg configurator.Configurator) *listenerBuilder {
//	return &listenerBuilder{
//		meshCatalog: meshCatalog,
//		svcAccount:  svcAccount,
//		cfg:         cfg,
//	}
//}

// Note: ServiceIdentity must be in the format "name.namespace" [https://github.com/openservicemesh/osm/issues/3188]
func newListenerBuilder(meshCatalog catalog.MeshCataloger, svcIdentity identity.ServiceIdentity, cfg configurator.Configurator, statsHeaders map[string]string) *listenerBuilder {
	return &listenerBuilder{
		meshCatalog:     meshCatalog,
		serviceIdentity: svcIdentity,
		cfg:             cfg,
		statsHeaders:    statsHeaders,
	}
}