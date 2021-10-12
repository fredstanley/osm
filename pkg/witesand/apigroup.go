package witesand

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func(wc *WitesandCatalog) FetchApigroupMap() map[string]ApigroupToPodMap {
	wc.Lock()
	defer wc.Unlock()
	return wc.apigroupToPodMap
}

func (wc *WitesandCatalog) UpdateApigroupMap(w http.ResponseWriter, r *http.Request) {
	var input map[string][]string
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil  {
		log.Error().Msgf("update [ApigroupMap] JSON decode err:%s", err)
		w.WriteHeader(400)
		fmt.Fprintf(w, "Decode error! please check your JSON formating.")
		return
	}

	wc.Lock()
	for apigroupName, pods := range input {
		atopMap := ApigroupToPodMap{
			Apigroup: apigroupName,
			Pods:     pods,
		}

		if apigroupName == "" {
			log.Error().Msgf("Error [ApigroupMap] ApigroupName is empty. name=%s pods=%+v", apigroupName, pods)
			continue
		}

		if len(atopMap.Pods) == 0 {
			// DELETE pods
			log.Info().Msgf("update [ApigroupMap] DELETE apigroup:%s", apigroupName)
			delete(wc.apigroupToPodMap, apigroupName)
			delete(wc.apigroupToPodIPMap, apigroupName)
		} else {
			// UPDATE pods
			log.Info().Msgf("[ApigroupMap] POST apigroup:%s pods:%+v", apigroupName, atopMap.Pods)
			wc.apigroupToPodMap[apigroupName] = atopMap
			wc.resolveApigroup(atopMap)
		}
	}
	wc.Unlock()

	wc.UpdateEnvoy()
}

func (wc *WitesandCatalog) UpdateAllApigroupMaps(apigroupToPodMap *map[string][]string) {
	wc.Lock()
	log.Info().Msgf("UpdateAll [ApigroupMap] updating %d apiggroups", len(*apigroupToPodMap))
	for apigroupName, pods := range *apigroupToPodMap {
		if apigroupName == "" {
			log.Error().Msgf("Error [ApigroupMap] ApigroupName is empty. name=%s pods=%+v", apigroupName, pods)
			continue
		}
		apigroupMap := ApigroupToPodMap{
			Apigroup: apigroupName,
			Pods:     pods,
		}
		wc.apigroupToPodMap[apigroupMap.Apigroup] = apigroupMap
	}
	wc.Unlock()

	wc.ResolveAllApigroups()
	wc.UpdateEnvoy()
}

// Resolve apigroup's pods to their respective IPs
func (wc *WitesandCatalog) resolveApigroup(atopmap ApigroupToPodMap) {
	atopipmap := ApigroupToPodIPMap{
		Apigroup: atopmap.Apigroup,
		PodIPs:   make([]string, 0),
	}
	for _, pod := range atopmap.Pods {
		podip := ""
		for _, podInfo := range wc.clusterPodMap {
			var exists bool
			if podip, exists = podInfo.PodToIPMap[pod]; exists {
				break
			}
		}
		if podip != "" {
			log.Info().Msgf("[ApigroupMap] RESOLVE pod:%s IP:%s", pod, podip)
			atopipmap.PodIPs = append(atopipmap.PodIPs, podip)
		} else {
			log.Info().Msgf("[ApigroupMap] CANNOT RESOLVE pod:%s !!", pod)
		}
	}
	wc.apigroupToPodIPMap[atopipmap.Apigroup] = atopipmap
}

func (wc *WitesandCatalog) ResolveAllApigroups() {
	wc.Lock()
	defer wc.Unlock()
	log.Info().Msgf("[ApigroupMap] Resovling all apigroups")
	for _, atopmap := range wc.apigroupToPodMap {
		wc.resolveApigroup(atopmap)
	}
}

func (wc *WitesandCatalog) ListApigroupClusterNames() ([]string, error) {
	wc.Lock()
	defer wc.Unlock()
	var apigroups []string
	for apigroup, _ := range wc.apigroupToPodMap {
		apigroups = append(apigroups, apigroup)
	}

	return apigroups, nil
}

func (wc *WitesandCatalog) ListApigroupToPodIPs() ([]ApigroupToPodIPMap, error) {
	wc.Lock()
	defer wc.Unlock()
	var atopipMaps []ApigroupToPodIPMap
	for _, atopipMap := range wc.apigroupToPodIPMap {
		atopipMaps = append(atopipMaps, atopipMap)
	}
	return atopipMaps, nil
}
