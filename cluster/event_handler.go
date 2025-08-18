package cluster

import (
	"context"
	"time"
)

const (
	defaultRemoteConfigReceiverTimeout = 5 * time.Second
)

type RemoteConfigReceiver interface {
	ReceiveConfig(ctx context.Context, remoteConfig []byte) error
}

type RoutesReceiver interface {
	ReceiveRoutes(ctx context.Context, routes RoutingConfig) error
}

type HostsUpgrader interface {
	Upgrade(hosts []string)
}

type EventHandler struct {
	remoteConfigReceiver RemoteConfigReceiver
	handleConfigTimeout  time.Duration
	routesReceiver       RoutesReceiver
	requiredModules      map[string]HostsUpgrader
}

func NewEventHandler() *EventHandler {
	return &EventHandler{
		requiredModules:     make(map[string]HostsUpgrader),
		handleConfigTimeout: defaultRemoteConfigReceiverTimeout,
	}
}

func (h *EventHandler) RemoteConfigReceiver(receiver RemoteConfigReceiver) *EventHandler {
	h.remoteConfigReceiver = receiver
	return h
}

func (h *EventHandler) RoutesReceiver(receiver RoutesReceiver) *EventHandler {
	h.routesReceiver = receiver
	return h
}

func (h *EventHandler) RequireModule(moduleName string, upgrader HostsUpgrader) *EventHandler {
	h.requiredModules[moduleName] = upgrader
	return h
}
