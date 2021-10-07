package catalog

import (
	"fmt"
	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/endpoint"
	"github.com/openservicemesh/osm/pkg/service"
	"github.com/openservicemesh/osm/pkg/witesand"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	//"time"
)
func (mc *MeshCatalog) GetProvider(ident string) endpoint.Provider {
	for _, ep := range mc.endpointsProviders {
		if ep.GetID() == ident {
			return ep
		}
	}
	return nil
}

func (mc *MeshCatalog) GetWitesandCataloger() witesand.WitesandCataloger {
	return mc.witesandCatalog
}

func (mc *MeshCatalog) GetWitesandCache(key string) ([]types.Resource, bool ) {
	if r, found := mc.witesandCatalog.Cache.Get(key); found {
		//<-time.After(1*time.Second)
		log.Error().Msgf( "cache len=%d", mc.witesandCatalog.Cache.Len())
		log.Error().Msgf( "cache keys=%+v", mc.witesandCatalog.Cache.Keys())
		return r.([]types.Resource), true
	}
	return nil, false
}

func (mc *MeshCatalog) SetWitesandCache(key string, result []types.Resource) bool {
	log.Error().Msgf( "cache set key=%+v", key)
	return mc.witesandCatalog.Cache.Set(key, result)
}

// ListLocalEndpoints returns the list of endpoints for this kubernetes cluster
func (mc *MeshCatalog) ListLocalClusterEndpoints() (map[string][]endpoint.Endpoint, error) {
	endpointMap := make(map[string][]endpoint.Endpoint)
	services := mc.kubeController.ListServices()
	for _, provider := range mc.endpointsProviders {
		if provider.GetID() != constants.KubeProviderName {
			continue
		}
		for _, svc := range services {
			log.Trace().Msgf("[ListLocalClusterEndpoints] service=%+v", svc.Name)
			meshSvc := service.MeshService {
				Namespace: "default",
				Name: svc.Name,
			}
			eps := provider.ListEndpointsForService(meshSvc)
			if len(eps) == 0 {
				continue
			}
			log.Trace().Msgf("[ListLocalClusterEndpoints] endpoints for service=%+v", eps)
			meshSvcStr := fmt.Sprintf("%s/%s", meshSvc.Namespace, meshSvc.Name)
			endpointMap[meshSvcStr] = eps
		}
	}
	return endpointMap, nil
}
