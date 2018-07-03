// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
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
	"log"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "install the container for <lang>. 'all' will install every possible language",
	Long:  `install the container for <lang>. 'all' will install every possible language`,
	Args:  cobra.MinimumNArgs(1),
	Run:   installCmdF,
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func installCmdF(cmd *cobra.Command, tags []string) {
	dockerClient, err := docker.NewClientFromEnv()
	if err != nil {
		log.Fatalf("Unnable to connect to docker: %v", err)
	}

	// TODO: HANDLE ERROR
	availableTagsList, _ := availableTags()
	for _, tag := range tags {
		if tag == "all" {
			for _, lang := range availableTagsList {
				installLanguage(dockerClient, lang)
			}
		} else if contains(availableTagsList, tag) {
			installLanguage(dockerClient, tag)
		} else {
			log.Fatalf("Language %s not a valid candidate", tag)
		}
	}
	return
}
