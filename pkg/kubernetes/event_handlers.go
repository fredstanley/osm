package kubernetes

import (
	"os"
	"reflect"

	"k8s.io/client-go/tools/cache"

	a "github.com/openservicemesh/osm/pkg/announcements"
	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/kubernetes/events"
	"github.com/openservicemesh/osm/pkg/metricsstore"
)

var emitLogs = os.Getenv(constants.EnvVarLogKubernetesEvents) == "true"

// observeFilter returns true for YES observe and false for NO do not pay attention to this
// This filter could be added optionally by anything using GetKubernetesEventHandlers()
type observeFilter func(obj interface{}) bool

// EventTypes is a struct helping pass the correct types to GetKubernetesEventHandlers
type EventTypes struct {
	Add    a.AnnouncementType
	Update a.AnnouncementType
	Delete a.AnnouncementType
}

// GetKubernetesEventHandlers creates Kubernetes events handlers.
func GetKubernetesEventHandlers(informerName, providerName string, shouldObserve observeFilter, eventTypes EventTypes) cache.ResourceEventHandlerFuncs {
	if shouldObserve == nil {
		shouldObserve = func(obj interface{}) bool { return true }
	}

//<<<<<<< HEAD
//	sendAnnouncement := func(eventType a.AnnouncementType, obj interface{}) {
//		if emitLogs {
//			log.Trace().Msgf("dispachertrack [%s][%s] %s event: %+v", providerName, informerName, eventType, obj)
//		}
//
//		if announce == nil {
//			return
//		}
//
//		ann := a.Announcement{
//			Type: eventType,
//		}
//
//		// getObjID is a function which has enough context to establish a
//		// ReferenceObjectID from the object for which this event occurred.
//		// For example the ReferenceObjectID for a Pod would be the Pod's UID.
//		// The getObjID function is optional;
//		if getObjID != nil {
//			ann.ReferencedObjectID = getObjID(obj)
//		}
//
//		select {
//		case announce <- ann:
//			// Channel post succeeded
//		default:
//			// Since pubsub introduction, there's a chance we start seeing full channels which
//			// will slowly become unused in favour of pub-sub subscriptions.
//			// We are making sure here ResourceEventHandlerFuncs never locks due to a push on a full channel.
//			//log.Trace().Msgf("Channel for provider %s is full, dropping channel notify %s ", providerName, eventType)
//		}
//	}
//
//=======
//>>>>>>> 865c66ed45ee888b5719d2e56a32f1534b61d1e7
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if !shouldObserve(obj) {
				logNotObservedNamespace(obj, eventTypes.Add, informerName, providerName)
				return
			}
			events.GetPubSubInstance().Publish(events.PubSubMessage{
				AnnouncementType: eventTypes.Add,
				NewObj:           obj,
				OldObj:           nil,
			})
			ns := getNamespace(obj)
			metricsstore.DefaultMetricsStore.K8sAPIEventCounter.WithLabelValues(eventTypes.Add.String(), ns).Inc()
			updateEventSpecificMetrics(eventTypes.Add)
		},

		UpdateFunc: func(oldObj, newObj interface{}) {
			if !shouldObserve(newObj) {
				logNotObservedNamespace(newObj, eventTypes.Update, informerName, providerName)
				return
			}
			events.GetPubSubInstance().Publish(events.PubSubMessage{
				AnnouncementType: eventTypes.Update,
				NewObj:           oldObj,
				OldObj:           newObj,
			})
			ns := getNamespace(newObj)
			metricsstore.DefaultMetricsStore.K8sAPIEventCounter.WithLabelValues(eventTypes.Update.String(), ns).Inc()
		},

		DeleteFunc: func(obj interface{}) {
			if !shouldObserve(obj) {
				logNotObservedNamespace(obj, eventTypes.Delete, informerName, providerName)
				return
			}
			events.GetPubSubInstance().Publish(events.PubSubMessage{
				AnnouncementType: eventTypes.Delete,
				NewObj:           nil,
				OldObj:           obj,
			})
			ns := getNamespace(obj)
			metricsstore.DefaultMetricsStore.K8sAPIEventCounter.WithLabelValues(eventTypes.Delete.String(), ns).Inc()
			updateEventSpecificMetrics(eventTypes.Delete)
		},
	}
}

func getNamespace(obj interface{}) string {
	return reflect.ValueOf(obj).Elem().FieldByName("ObjectMeta").FieldByName("Namespace").String()
}

func logNotObservedNamespace(obj interface{}, eventType a.AnnouncementType, informerName, providerName string) {
	if emitLogs {
		log.Debug().Msgf("Namespace %q is not observed by OSM; ignoring %s event; informer=%s; provider=%s", getNamespace(obj), eventType, informerName, providerName)
	}
}
