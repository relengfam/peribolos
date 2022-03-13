/*
Copyright RelEngFam Authors.

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

package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"k8s.io/test-infra/prow/config/org"

	"github.com/relengfam/peribolos/options"
)

func Run(o *options.Options) error {
	githubClient, err := o.GithubOpts.GitHubClient(!o.Confirm)
	if err != nil {
		logrus.WithError(err).Fatal("Error getting GitHub client.")
	}

	if o.Dump != "" {
		ret, err := dumpOrgConfig(githubClient, o.Dump, o.IgnoreSecretTeams)
		if err != nil {
			logrus.WithError(err).Fatalf("Dump %s failed to collect current data.", o.Dump)
		}
		var output interface{}
		if o.DumpFull {
			output = org.FullConfig{
				Orgs: map[string]org.Config{o.Dump: *ret},
			}
		} else {
			output = ret
		}
		out, err := yaml.Marshal(output)
		if err != nil {
			logrus.WithError(err).Fatalf("Dump %s failed to marshal output.", o.Dump)
		}

		logrus.Infof("Dumping orgs[\"%s\"]:", o.Dump)
		fmt.Println(string(out))

		return nil
	}

	raw, err := ioutil.ReadFile(o.Config)
	if err != nil {
		logrus.WithError(err).Fatal("Could not read --config-path file")
	}

	var cfg org.FullConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		logrus.WithError(err).Fatal("Failed to load configuration")
	}

	for name, orgcfg := range cfg.Orgs {
		if err := configureOrg(*o, githubClient, name, orgcfg); err != nil {
			logrus.Fatalf("Configuration failed: %v", err)
		}
	}

	logrus.Info("Finished syncing configuration.")

	return nil
}
