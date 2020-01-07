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
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestRead(t *testing.T) {
	type args struct {
		path string
		cfg  interface{}
	}
	type test struct {
		name      string
		args      args
		checkFunc func(interface{}) error
		wantErr   bool
	}
	tests := []test{
		func() test {
			type dummyConfig struct {
				dummyValue string
			}
			return test{
				name: "read yaml success",
				args: args{
					path: "./testdata/dummyConfig.yaml",
					cfg:  &dummyConfig{},
				},
				checkFunc: func(cfg interface{}) error {
					c, ok := cfg.(*dummyConfig)
					if !ok {
						return errors.New("type cast error")
					}
					if c.dummyValue != "dummy" {
						return fmt.Errorf("dummyValue incorrect, got: %s", c.dummyValue)
					}
					return nil
				},
				wantErr: false,
			}
		}(),
		func() test {
			type dummyConfig struct {
				dummyValue string
			}
			return test{
				name: "read json success",
				args: args{
					path: "./testdata/dummyConfig.json",
					cfg:  &dummyConfig{},
				},
				checkFunc: func(cfg interface{}) error {
					c, ok := cfg.(*dummyConfig)
					if !ok {
						return errors.New("type cast error")
					}
					if c.dummyValue != "dummy" {
						return fmt.Errorf("dummyValue incorrect, got: %s", c.dummyValue)
					}
					return nil
				},
				wantErr: false,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Read(tt.args.path, tt.args.cfg); (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := tt.checkFunc(tt.args.cfg); err != nil {
				t.Errorf("Read() error = %v", err)
			}
		})
	}
}

func TestGetActualValue(t *testing.T) {
	type args struct {
		val string
	}
	tests := []struct {
		name       string
		args       args
		beforeTest func()
		afterTest  func()
		want       string
	}{
		{
			name: "val is not env var",
			args: args{
				val: "dummy",
			},
			want: "dummy",
		},
		{
			name: "val is env var",
			args: args{
				val: "_getActualValueTest_",
			},
			beforeTest: func() {
				os.Setenv("getActualValueTest", "dummyEnvVal")
			},
			want: "dummyEnvVal",
			afterTest: func() {
				os.Unset("getActualValueTest")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeTest != nil {
				tt.beforeTest()
			}
			if tt.afterTest != nil {
				defer tt.afterTest()
			}

			if got := GetActualValue(tt.args.val); got != tt.want {
				t.Errorf("GetActualValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetActualValues(t *testing.T) {
	type args struct {
		vals []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetActualValues(tt.args.vals); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetActualValues() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToRawYaml(t *testing.T) {
	type args struct {
		data interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "ToRawYaml return the yaml string",
			args: args{
				data: struct {
					dummyData1 string
					dummyData2 int
				}{
					dummyData1: "dummyString",
					dummyData2: 1,
				},
			},
			want: `dummyData1: dummyString
dummyData2: 2`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToRawYaml(tt.args.data); got != tt.want {
				t.Errorf("ToRawYaml() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isEnvVar(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "isEnvVar return true",
			args: args{
				str: "_dummyStr_",
			},
			want: true,
		},
		{
			name: "isEnvVar return false",
			args: args{
				str: "dummyStr",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isEnvVar(tt.args.str); got != tt.want {
				t.Errorf("isEnvVar() = %v, want %v", got, tt.want)
			}
		})
	}
}
