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

// Package compress provides compress functions
package compress

import (
	"bytes"
	"encoding/gob"
	"reflect"

	"github.com/vdaas/vald/internal/errors"
)

type gobCompressor struct {
}

func NewGob(opts ...GobOption) (Compressor, error) {
	c := new(gobCompressor)
	for _, opt := range append(defaultGobOpts, opts...) {
		if err := opt(c); err != nil {
			return nil, errors.ErrOptionFailed(err, reflect.ValueOf(opt))
		}
	}

	return c, nil
}

func (g *gobCompressor) CompressVector(vector []float32) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := gob.NewEncoder(buf).Encode(vector)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (g *gobCompressor) DecompressVector(bs []byte) ([]float32, error) {
	var vector []float32
	err := gob.NewDecoder(bytes.NewBuffer(bs)).Decode(&vector)
	if err != nil {
		return nil, err
	}

	return vector, nil
}
