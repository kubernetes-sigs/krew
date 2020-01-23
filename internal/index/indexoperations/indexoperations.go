// Copyright 2019 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package indexoperations

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/internal/gitutil"
)

type IndexConfig struct {
	Indices map[string]string `yaml:"indices"`
}

func (i IndexConfig) AddIndex(alias, uri string) error {
	i.Indices[alias] = uri
	err := gitutil.EnsureUpdated(uri, environment.MustGetKrewPaths().CustomIndicesPath(alias))
	if err != nil {
		return err
	}
	return createIndexConfigFile(&i)
}

func (i *IndexConfig) RemoveIndex(key string) error {
	if _, ok := i.Indices[key]; !ok {
		return fmt.Errorf("Must provide a valid index name to remove")
	}
	delete(i.Indices, key)
	return createIndexConfigFile(i)
}

func GetIndexConfig() (*IndexConfig, error) {
	config, err := ioutil.ReadFile(environment.MustGetKrewPaths().IndexConfigPath() + "/indexconfig.yaml")
	if os.IsNotExist(err) {
		indexConfig := &IndexConfig{Indices: make(map[string]string, 0)}
		err = createIndexConfigFile(indexConfig)
		return indexConfig, err
	} else if err != nil {
		return nil, err
	}
	indexConfig := &IndexConfig{Indices: make(map[string]string, 0)}
	err = yaml.Unmarshal(config, indexConfig)
	if err != nil {
		return nil, err
	}
	return indexConfig, nil
}

func createIndexConfigFile(indexConfig *IndexConfig) error {
	out, err := yaml.Marshal(indexConfig)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(environment.MustGetKrewPaths().IndexConfigPath()+"/indexconfig.yaml", out, 0644)
}
