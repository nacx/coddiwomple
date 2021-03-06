// Copyright 2018 Tetrate, Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/istio-ecosystem/coddiwomple/pkg/datamodel"
	"github.com/istio-ecosystem/coddiwomple/pkg/datamodel/mem"
)

type services []datamodel.GlobalService

func serviceFromFile(path string) (datamodel.DataModel, error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open file %q", path)
	}
	gss := services{}
	if err := json.Unmarshal(contents, &gss); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal file as json")
	}

	dm := mem.NewDataModel()
	for _, gs := range gss {
		svc := gs
		dm.CreateGlobalService(&svc)
	}
	return dm, nil
}

// wrapper type which adds fields we want in the CLI representation but not the datamodel.
type cluster struct {
	datamodel.Cluster

	KubeconfigPath    string `json:"kubeconfig_path"`
	KubeconfigContext string `json:"kubeconfig_context"`
}

func clustersFromFile(path string) ([]string, []cluster, datamodel.Infrastructure, error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return []string{}, []cluster{}, nil, errors.Wrapf(err, "could not open file %q", path)
	}
	c := make([]cluster, 0, 10)
	if err := json.Unmarshal(contents, &c); err != nil {
		return []string{}, []cluster{}, nil, errors.Wrap(err, "could not unmarshal file as json")
	}

	names := make([]string, len(c))
	cls := make(map[string]string, len(c))
	for i, cl := range c {
		names[i] = cl.Name
		cls[cl.Name] = cl.Address
	}
	sort.Strings(names)
	return names, c, mem.Infrastructure(cls), nil
}

func clustersFlagToInfra(clusters []string) ([]string, datamodel.Infrastructure, error) {
	cls := make(map[string]string, len(clusters))
	names := make([]string, 0, len(clusters))
	var errs error
	for i, c := range clusters {
		parts := strings.Split(c, ":")
		if len(parts) != 2 {
			errs = multierror.Append(errs, fmt.Errorf("expected `name:address` pairs but got %q", c))
			continue
		}
		cls[parts[0]] = parts[1]
		names[i] = parts[0]
	}
	sort.Strings(names)
	return names, mem.Infrastructure(cls), errs
}
