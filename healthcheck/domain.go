package healthcheck

import (
	"time"
)

const (
	StatusPass = "pass"
	StatusFail = "fail"
)

// nolint:tagliatelle
type Detail struct {
	ComponentName string
	ComponentType string
	Status        string
	Output        string `json:",omitempty"`
	Time          time.Time
}

type Result struct {
	Status  string
	Details map[string][]Detail
}
