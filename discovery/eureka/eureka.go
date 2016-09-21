package eureka

import (
	"fmt"
	"sort"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/joyent/containerpilot/discovery"
	eurekaclient "github.com/joyent/containerpilot/discovery/eureka/eurekaclient"
)

func init() {
	discovery.RegisterBackend("eureka", ConfigHook)
}

// Eureka is a service discovery backend from Netflix OSS
type Eureka struct{ eurekaclient.Client }

// ConfigHook is the hook to register with the Eurkea backend
func ConfigHook(raw interface{}) (discovery.ServiceBackend, error) {
	return NewEurekaConfig(raw)
}

// NewEurekaConfig creates a new service discovery backend for Eureka
func NewEurekaConfig(config interface{}) (*Eureka, error) {
	var machines []string
	var err error
	switch t := config.(type) {
	case string:
		machines = []string{t}
	case []string:
		machines = t
	default:
		return nil, fmt.Errorf("Unexpected Eureka config structure. Expected a string or an array of strings")
	}
	if err != nil {
		return nil, err
	}

	client := eurekaclient.NewClient(machines)
	return &Eureka{*client}, nil
}

// Deregister removes the node from Eureka.
func (e *Eureka) Deregister(service *discovery.ServiceDefinition) {
	e.MarkForMaintenance(service)
}

// MarkForMaintenance removes the node from Eureka.
func (e *Eureka) MarkForMaintenance(service *discovery.ServiceDefinition) {
	if err := e.UnregisterInstance(service.Name, cleanID(service.ID)); err != nil {
		log.Infof("Deregistering failed: %s", err)
	}
}

// SendHeartbeat triggers a heartbeat against Eureka.
func (e *Eureka) SendHeartbeat(service *discovery.ServiceDefinition) {
	_, err := e.GetInstance(service.Name, cleanID(service.ID))
	if err != nil {
		log.Infof("Registering service %s in Eureka %s:%v", service.Name, service.IPAddress, service.Port)
		if err = e.registerService(*service); err != nil {
			log.Warnf("Service registration failed: %s", err)
		}
	}
	if err = e.Client.SendHeartbeat(service.Name, cleanID(service.ID)); err != nil {
		log.Warnf("Heartbeat failed: %s", err)
	}
}

func (e *Eureka) registerService(service discovery.ServiceDefinition) error {
	return e.RegisterInstance(service.Name,
		eurekaclient.NewInstanceInfo(
			cleanID(service.ID),
			service.Name,
			service.Name,
			service.IPAddress,
			service.Port,
			uint(service.TTL),
			false),
	)
}

var upstreams = make(map[string][]eurekaclient.InstanceInfo)

// CheckForUpstreamChanges runs the health check
func (e Eureka) CheckForUpstreamChanges(backendName, backendTag string) bool {
	app, err := e.GetApplication(backendName)
	if err != nil {
		log.Warnf("Failed to query %v: %s", backendName, err)
		return false
	}
	didChange := compareForChange(upstreams[backendName], app.Instances)
	if didChange || len(app.Instances) == 0 {
		// We don't want to cause an onChange event the first time we read-in
		// but we do want to make sure we've written the key for this map
		upstreams[backendName] = app.Instances
	}
	return didChange
}

// Compare the two arrays to see if the address or port has changed
// or if we've added or removed entries.
func compareForChange(existing, new []eurekaclient.InstanceInfo) bool {

	if len(existing) != len(new) {
		return true
	}

	sort.Sort(ByIPAddr(existing))
	sort.Sort(ByIPAddr(new))
	for i, ex := range existing {
		if ex.IpAddr != new[i].IpAddr || ex.Port.Port != new[i].Port.Port {
			return true
		}
	}
	return false
}

// Returns the uri broken into an address and scheme portion
func cleanID(raw string) string {

	var s = raw
	s = strings.Replace(s, ".", "-", -1)
	s = strings.ToLower(s)
	return s

}

// ByIPAddr implements the Sort interface because Go can't sort without it.
type ByIPAddr []eurekaclient.InstanceInfo

func (se ByIPAddr) Len() int      { return len(se) }
func (se ByIPAddr) Swap(i, j int) { se[i], se[j] = se[j], se[i] }
func (se ByIPAddr) Less(i, j int) bool {
	return se[i].IpAddr < se[j].IpAddr
}
