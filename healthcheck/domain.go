package healthcheck

import (
	"time"
)

const (
	StatusPass = "pass"
	StatusFail = "fail"
)

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
