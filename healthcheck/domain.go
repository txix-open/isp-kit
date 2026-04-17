// Package healthcheck provides utilities for implementing and managing health checks
// in applications. It allows registering component checks, caching results, and
// exposing them via an HTTP endpoint in JSON format compliant with the
// draft-inadarei-api-health-check specification.
package healthcheck

import (
	"time"
)

const (
	// StatusPass indicates a healthy component.
	StatusPass = "pass"
	// StatusFail indicates an unhealthy component.
	StatusFail = "fail"
)

// Detail represents the health status of a single component.
//
// nolint:tagliatelle
type Detail struct {
	// componentName is the name of the component being checked.
	ComponentName string
	// componentType describes the type of the component.
	ComponentType string
	// status indicates whether the component is healthy ("pass") or unhealthy ("fail").
	Status string
	// output contains additional information about the health check, such as error messages.
	Output string `json:",omitempty"`
	// time is the timestamp when the health check was performed.
	Time time.Time
}

// Result represents the aggregate health status of all registered components.
type Result struct {
	// status is the overall health status ("pass" if all components are healthy, "fail" otherwise).
	Status string
	// details maps component names to their individual health details.
	Details map[string][]Detail
}
