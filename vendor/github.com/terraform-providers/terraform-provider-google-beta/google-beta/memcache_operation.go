// ----------------------------------------------------------------------------
//
//     ***     AUTO GENERATED CODE    ***    AUTO GENERATED CODE     ***
//
// ----------------------------------------------------------------------------
//
//     This file is automatically generated by Magic Modules and manual
//     changes will be clobbered when the file is regenerated.
//
//     Please read more about how to change this file in
//     .github/CONTRIBUTING.md.
//
// ----------------------------------------------------------------------------
package google

import (
	"encoding/json"
	"fmt"
	"time"
)

type MemcacheOperationWaiter struct {
	Config  *Config
	Project string
	CommonOperationWaiter
}

func (w *MemcacheOperationWaiter) QueryOp() (interface{}, error) {
	if w == nil {
		return nil, fmt.Errorf("Cannot query operation, it's unset or nil.")
	}
	// Returns the proper get.
	url := fmt.Sprintf("https://memcache.googleapis.com/v1beta2/%s", w.CommonOperationWaiter.Op.Name)
	return sendRequest(w.Config, "GET", w.Project, url, nil)
}

func createMemcacheWaiter(config *Config, op map[string]interface{}, project, activity string) (*MemcacheOperationWaiter, error) {
	if val, ok := op["name"]; !ok || val == "" {
		// This was a synchronous call - there is no operation to wait for.
		return nil, nil
	}
	w := &MemcacheOperationWaiter{
		Config:  config,
		Project: project,
	}
	if err := w.CommonOperationWaiter.SetOp(op); err != nil {
		return nil, err
	}
	return w, nil
}

// nolint: deadcode,unused
func memcacheOperationWaitTimeWithResponse(config *Config, op map[string]interface{}, response *map[string]interface{}, project, activity string, timeout time.Duration) error {
	w, err := createMemcacheWaiter(config, op, project, activity)
	if err != nil || w == nil {
		// If w is nil, the op was synchronous.
		return err
	}
	if err := OperationWait(w, activity, timeout, config.PollInterval); err != nil {
		return err
	}
	return json.Unmarshal([]byte(w.CommonOperationWaiter.Op.Response), response)
}

func memcacheOperationWaitTime(config *Config, op map[string]interface{}, project, activity string, timeout time.Duration) error {
	w, err := createMemcacheWaiter(config, op, project, activity)
	if err != nil || w == nil {
		// If w is nil, the op was synchronous.
		return err
	}
	return OperationWait(w, activity, timeout, config.PollInterval)
}
