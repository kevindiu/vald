//
// Copyright (C) 2019-2020 Vdaas.org Vald team ( kpango, rinx, kmrmt )
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// Package service manages the main logic of server.
package service

import (
	"context"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/vdaas/vald/internal/errgroup"
	"github.com/vdaas/vald/internal/errors"
	"github.com/vdaas/vald/internal/k8s"
	"github.com/vdaas/vald/internal/k8s/pod"
	"github.com/vdaas/vald/internal/log"
	"github.com/vdaas/vald/internal/safety"
)

type Replicator interface {
	Start(context.Context) (<-chan error, error)
	GetDeletedPods() ([]string, bool)
}

type replicator struct {
	pods      atomic.Value
	ctrl      k8s.Controller
	namespace string
	name      string
	csd       time.Duration
	eg        errgroup.Group
}

func New(opts ...Option) (rp Replicator, err error) {
	r := new(replicator)
	for _, opt := range append(defaultOpts, opts...) {
		if err := opt(r); err != nil {
			return nil, errors.ErrOptionFailed(err, reflect.ValueOf(opt))
		}
	}
	r.pods.Store(make([]string, 0, 0))

	r.ctrl, err = k8s.New(
		k8s.WithControllerName("vald k8s replication manager controller"),
		k8s.WithEnableLeaderElection(),
		k8s.WithResourceController(pod.New(
			pod.WithControllerName("pod discoverer"),
			pod.WithOnErrorFunc(func(err error) {
				log.Error(err)
			}),
			pod.WithOnReconcileFunc(func(podList map[string][]pod.Pod) {
				log.Debugf("pod resource reconciled\t%#v", podList)
			}),
		)),
	)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *replicator) Start(ctx context.Context) (<-chan error, error) {
	rech, err := r.ctrl.Start(ctx)
	if err != nil {
		return nil, err
	}
	ech := make(chan error, 2)
	r.eg.Go(safety.RecoverFunc(func() (err error) {
		defer close(ech)
		rt := time.NewTicker(r.csd)
		defer rt.Stop()
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-rt.C:

			case err = <-rech:
				if err != nil {
					ech <- err
				}
			}

		}
	}))
	return ech, nil
}

func (r *replicator) GetDeletedPods() ([]string, bool) {
	return nil, false
}
