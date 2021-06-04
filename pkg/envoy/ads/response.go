package ads

import (
	"strconv"
	"time"

	xds_discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"

	"github.com/openservicemesh/osm/pkg/configurator"
	"github.com/openservicemesh/osm/pkg/envoy"
)

const (
	// ADSUpdateStr is a constant string value to identify full XDS update times on metric labels
	ADSUpdateStr = "ADS"
)

// Wrapper to create and send a discovery response to an envoy server
func (s *Server) sendTypeResponse(tURI envoy.TypeURI,
	proxy *envoy.Proxy, server *xds_discovery.AggregatedDiscoveryService_StreamAggregatedResourcesServer,
	req *xds_discovery.DiscoveryRequest, cfg configurator.Configurator) error {
	// Tracks the success of this TypeURI response operation; accounts also for receipt on envoy server side
	success := false
	xdsShortName := envoy.XDSShortURINames[tURI]
	defer xdsPathTimeTrack(time.Now(), xdsShortName, proxy.GetCertificateCommonName().String(), &success)

	log.Trace().Msgf("[%s] Creating response for proxy with SerialNumber=%s on Pod with UID=%s", xdsShortName, proxy.GetCertificateSerialNumber(), proxy.GetPodUID())

	discoveryResponse, err := s.newAggregatedDiscoveryResponse(proxy, req, cfg)
	if err != nil {
		log.Error().Err(err).Msgf("[%s] Failed to create response for proxy with SerialNumber=%s on Pod with UID=%s", xdsShortName, proxy.GetCertificateSerialNumber(), proxy.GetPodUID())
		return err
	}

	if err := (*server).Send(discoveryResponse); err != nil {
		log.Error().Err(err).Msgf("[%s] Error sending to proxy with SerialNumber=%s on Pod with UID=%s", xdsShortName, proxy.GetCertificateSerialNumber(), proxy.GetPodUID())
		return err
	}

	success = true // read by deferred function
	return nil
}

func (s *Server) sendAllResponses(proxy *envoy.Proxy, server *xds_discovery.AggregatedDiscoveryService_StreamAggregatedResourcesServer, cfg configurator.Configurator) {
	log.Trace().Msgf("A change announcement triggered *DS update for proxy with SerialNumber=%s on Pod with UID=%s", proxy.GetCertificateSerialNumber(), proxy.GetPodUID())

	// Tracks the success of this full update of all its XDS paths. If a single XDS response path fails for this full update,
	// the full updated will be considered as failed for metric purposes (success = false)
	success := true
	defer xdsPathTimeTrack(time.Now(), ADSUpdateStr, proxy.GetCertificateCommonName().String(), &success)

	// Order is important: CDS, EDS, LDS, RDS
	// See: https://github.com/envoyproxy/go-control-plane/issues/59
	for _, typeURI := range envoy.XDSResponseOrder {
		// For SDS we need to add ResourceNames
		var request *xds_discovery.DiscoveryRequest
		if typeURI == envoy.TypeSDS {
			request = makeRequestForAllSecrets(proxy, s.catalog)
			if request == nil {
				continue
			}
		} else {
			request = &xds_discovery.DiscoveryRequest{TypeUrl: string(typeURI)}
		}

		err := s.sendTypeResponse(typeURI, proxy, server, request, cfg)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to create and send %s update to Proxy %s",
				envoy.XDSShortURINames[typeURI], proxy.GetCertificateCommonName())
			success = false
		}
	}
}

// sendSDSResponse sends an SDS response to the given proxy containing all associated secrets
func (s *Server) sendSDSResponse(proxy *envoy.Proxy, server *xds_discovery.AggregatedDiscoveryService_StreamAggregatedResourcesServer, cfg configurator.Configurator) {
	request := makeRequestForAllSecrets(proxy, s.catalog)

//<<<<<<< HEAD
//	// There is an SDS validation cert corresponding to each upstream service
//	upstreamServices, err := meshCatalog.ListAllowedOutboundServices(serviceForProxy)
//	if err != nil {
//		log.Error().Err(err).Msgf("Error listing outbound services for proxy %q", serviceForProxy)
//		return nil
//	}
//	for _, upstream := range upstreamServices {
//		upstreamRootCertResource := envoy.SDSCert{
//			MeshService: upstream.GetMeshService(),
//			CertType:    envoy.RootCertTypeForMTLSOutbound,
//		}.String()
//		discoveryRequest.ResourceNames = append(discoveryRequest.ResourceNames, upstreamRootCertResource)
//=======
	if err := s.sendTypeResponse(envoy.TypeSDS, proxy, server, request, cfg); err != nil {
		log.Error().Err(err).Msgf("Failed to create and send %s update to Proxy %s",
			envoy.TypeSDS, proxy.GetCertificateCommonName())
//>>>>>>> 3d923b3f2d72006f6cdaad056938c492c364196d
	}
}

func (s *Server) newAggregatedDiscoveryResponse(proxy *envoy.Proxy, request *xds_discovery.DiscoveryRequest, cfg configurator.Configurator) (*xds_discovery.DiscoveryResponse, error) {
	typeURL := envoy.TypeURI(request.TypeUrl)
	handler, ok := s.xdsHandlers[typeURL]
	if !ok {
		log.Error().Msgf("Responder for TypeUrl %s is not implemented", request.TypeUrl)
		return nil, errUnknownTypeURL
	}

	if s.cfg.IsDebugServerEnabled() {
		s.trackXDSLog(proxy.GetCertificateCommonName(), typeURL)
	}

	// request.Node is only available on the first Discovery Request; will be nil on the following
	nodeID := ""
	if request.Node != nil {
		nodeID = request.Node.Id
	}

	log.Trace().Msgf("Invoking handler for type %s; request from Envoy with Node ID %s", typeURL, nodeID)
	response, err := handler(s.catalog, proxy, request, cfg, s.certManager)
	if err != nil {
		log.Error().Msgf("Responder for TypeUrl %s is not implemented", request.TypeUrl)
		return nil, errCreatingResponse
	}

	response.Nonce = proxy.SetNewNonce(typeURL)
	response.VersionInfo = strconv.FormatUint(proxy.IncrementLastSentVersion(typeURL), 10)

//<<<<<<< HEAD
//	//if envoy.TypeURI(request.TypeUrl) == envoy.TypeSDS {
//	//	log.Trace().Msgf("Constructed %s response: VersionInfo=%s", response.TypeUrl, response.VersionInfo)
//	//} else {
//	//	log.Trace().Msgf("Constructed %s response: VersionInfo=%s; %+v", response.TypeUrl, response.VersionInfo, response)
//	//}
//=======
	// NOTE: Never log entire 'response' - will contain secrets!
	log.Trace().Msgf("Constructed %s response: VersionInfo=%s", response.TypeUrl, response.VersionInfo)
//>>>>>>> 3d923b3f2d72006f6cdaad056938c492c364196d

	return response, nil
}
