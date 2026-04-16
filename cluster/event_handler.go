package cluster

import (
	"context"
	"time"
)

// RemoteConfigReceiver is an interface for receiving and processing remote configuration updates.
type RemoteConfigReceiver interface {
	// ReceiveConfig processes the remote configuration.
	ReceiveConfig(ctx context.Context, remoteConfig []byte) error
}

// RoutesReceiver is an interface for receiving and processing routing configuration updates.
type RoutesReceiver interface {
	// ReceiveRoutes processes the routing configuration.
	ReceiveRoutes(ctx context.Context, routes RoutingConfig) error
}

// HostsUpgrader is an interface for upgrading or updating host lists.
type HostsUpgrader interface {
	// Upgrade updates the list of hosts.
	Upgrade(hosts []string)
}

const (
	defaultRemoteConfigReceiverTimeout = 5 * time.Second
)

// EventHandler manages event handling for remote configuration, routes, and module dependencies.
type EventHandler struct {
	remoteConfigReceiver RemoteConfigReceiver
	handleConfigTimeout  time.Duration
	routesReceiver       RoutesReceiver
	requiredModules      map[string]HostsUpgrader
}

// NewEventHandler creates a new EventHandler with default settings.
func NewEventHandler() *EventHandler {
	return &EventHandler{
		requiredModules:     make(map[string]HostsUpgrader),
		handleConfigTimeout: defaultRemoteConfigReceiverTimeout,
	}
}

// RemoteConfigReceiver sets the receiver for remote configuration updates and returns
// the EventHandler for method chaining.
func (h *EventHandler) RemoteConfigReceiver(receiver RemoteConfigReceiver) *EventHandler {
	h.remoteConfigReceiver = receiver
	return h
}

// RoutesReceiver sets the receiver for routing configuration updates and returns
// the EventHandler for method chaining.
func (h *EventHandler) RoutesReceiver(receiver RoutesReceiver) *EventHandler {
	h.routesReceiver = receiver
	return h
}

// RequireModule registers a module dependency with its corresponding hosts upgrader
// and returns the EventHandler for method chaining.
func (h *EventHandler) RequireModule(moduleName string, upgrader HostsUpgrader) *EventHandler {
	h.requiredModules[moduleName] = upgrader
	return h
}
