/*
Copyright 2019 The Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	crdv1 "github.com/moolen/harbor-sync/api/v1"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

// Adapter waits for events on an input channel
// and emits SyncConfig change Events using a GenericEvent on a different channel
type Adapter struct {
	client.Client
	Input []<-chan struct{}
}

// NewAdapter creates a new adapter instance
func NewAdapter(c client.Client, input []<-chan struct{}) Adapter {
	return Adapter{
		c,
		input,
	}
}

// Run starts a goroutine which reads from the input channel
// and writes SyncConfig change events to the returned channel
func (a Adapter) Run() <-chan event.GenericEvent {
	c := make(chan event.GenericEvent)
	for _, input := range a.Input {
		go a.loop(c, input)
	}
	return c
}

// loop reads from the input channel and puts GenericEvents into the output chan
// FIXME: add stop channel
func (a Adapter) loop(c chan event.GenericEvent, input <-chan struct{}) {
	for {
		select {
		case <-input:
			log.WithFields(log.Fields{
				"component": "adapter",
				"action":    "loop",
			}).Debugf("received reconcile event from poller")
			ctx := context.Background()
			var cfgs crdv1.HarborSyncList
			err := a.List(ctx, &cfgs)
			if err != nil {
				log.WithFields(log.Fields{
					"component": "adapter",
					"action":    "loop",
				}).Errorf("error fetching harbor sync config: %s", err)
				continue
			}
			// emit events using generic Event
			for i := range cfgs.Items {
				c <- event.GenericEvent{
					Object: &cfgs.Items[i],
				}
			}
		}
	}
}
