//
// Copyright (C) 2019 Vdaas.org Vald team ( kpango, kou-m, rinx )
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// Package service
package service

import "github.com/vdaas/vald/internal/net/grpc"

type FilterOption func(f *filter) error

var (
	defaultFilterOpts = []FilterOption{}
)

func WithFilterClient(client grpc.Client) FilterOption {
	return func(f *filter) error {
		if client != nil {
			f.client = client
		}
		return nil
	}
}