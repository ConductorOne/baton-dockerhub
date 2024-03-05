/*
   Copyright 2020 Docker Hub Tool authors

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

package external

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/cli/cli/command"
	dockercredentials "github.com/docker/cli/cli/config/credentials"
	cliflags "github.com/docker/cli/cli/flags"
)

func GetCredentials(ctx context.Context) (string, string, error) {
	dockerCli, err := command.NewDockerCli()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return "", "", fmt.Errorf("failed to create docker client: %w", err)
	}
	opts := cliflags.NewClientOptions()
	if err := dockerCli.Initialize(opts); err != nil {
		return "", "", fmt.Errorf("failed to initialize docker client: %w", err)
	}

	store := NewStore(func(key string) dockercredentials.Store {
		config := dockerCli.ConfigFile()
		return config.GetCredentialsStore(key)
	})
	auth, err := store.GetAuth()
	if err != nil {
		return "", "", fmt.Errorf("failed to get auth: %w", err)
	}

	return auth.Token, auth.RefreshToken, nil
}
