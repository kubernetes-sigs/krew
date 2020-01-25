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
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"

	"sigs.k8s.io/krew/internal/environment"
	"sigs.k8s.io/krew/internal/gitutil"
	"sigs.k8s.io/krew/pkg/constants"
)

type IndexConfig struct {
	Indices map[string]string `yaml:"indices"`
}

func (i IndexConfig) AddIndex(alias, uri string) error {
	err := gitutil.EnsureUpdated(uri, environment.MustGetKrewPaths().CustomIndexPath(alias))
	if err != nil {
		return err
	}
	i.Indices[alias] = uri
	return createIndexConfigFile(&i)
}

func (i *IndexConfig) RemoveIndex(key string) error {
	if _, ok := i.Indices[key]; !ok {
		return errors.Errorf("must provide a valid index name to remove")
	}
	err := os.RemoveAll(environment.MustGetKrewPaths().CustomIndexPath(key))
	if err != nil {
		return err
	}
	delete(i.Indices, key)
	return createIndexConfigFile(i)
}

func GetIndexConfig() (*IndexConfig, error) {
	config, err := ioutil.ReadFile(environment.MustGetKrewPaths().IndexConfigPath() + "/indexconfig" + constants.ManifestExtension)
	if os.IsNotExist(err) {
		indexConfig := &IndexConfig{Indices: make(map[string]string)}
		err = createIndexConfigFile(indexConfig)
		return indexConfig, err
	} else if err != nil {
		return nil, err
	}
	indexConfig := &IndexConfig{Indices: make(map[string]string)}
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
	return ioutil.WriteFile(environment.MustGetKrewPaths().IndexConfigPath()+"/indexconfig"+constants.ManifestExtension, out, 0644)
}
