//
// Copyright (C) 2019 Vdaas.org Vald team ( kpango, kmrmt, rinx )
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

// package config providers configuration type and load configuration logic
package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/vdaas/vald/internal/net/http/json"
	yaml "gopkg.in/yaml.v2"
)

const (
	envVarIndicator = "_"
)

// Read reads the file located in path and decode the corresponding yaml or json file to cfg struct.
func Read(path string, cfg interface{}) error {
	f, err := os.OpenFile(path, os.O_RDONLY, 0600)
	if err != nil {
		return err
	}
	switch filepath.Ext(path) {
	case ".yaml":
		err = yaml.NewDecoder(f).Decode(cfg)
	case ".json":
		err = json.Decode(f, cfg)
	}
	return err
}

// GetActualValue returns the environment variable value if the val has prefix and suffix "_", otherwise the val will directly return.
func GetActualValue(val string) string {
	if isEnvVar(val) {
		return os.ExpandEnv(os.Getenv(strings.TrimPrefix(strings.TrimSuffix(val, "_"), "_")))
	}
	return os.ExpandEnv(val)
}

// GetActualValues is the same as GetActualValue, but it process a string slice.
func GetActualValues(vals []string) []string {
	result := make([]string, len(vals))
	for i, val := range vals {
		result[i] = GetActualValue(val)
	}
	return result
}

// ToRawYaml returns the encoded yaml string from the data.
func ToRawYaml(data interface{}) string {
	buf := bytes.NewBuffer(nil)
	yaml.NewEncoder(buf).Encode(data)
	return buf.String()
}

// isEnvVar returns if the str contains "_" prefix and suffix.
func isEnvVar(str string) bool {
	return strings.HasPrefix(str, envVarIndicator) && strings.HasSuffix(str, envVarIndicator)
}
