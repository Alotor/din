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
	"fmt"
	"log"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update a language container",
	Long:  `update a language container`,
	Run:   updateCmdF,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func updateCmdF(cmd *cobra.Command, tags []string) {
	dockerClient, err := docker.NewClientFromEnv()
	if err != nil {
		log.Fatalf("Unnable to connect to docker: %v", err)
	}

	// TODO: HANDLE ERROR
	availableTagsList, _ := availableTags()
	if len(tags) == 0 && confirm("update all images") {
		for _, tag := range availableTagsList {
			if checkImageExists(dockerClient, tag) {
				tags = append(tags, tag)
			}
		}
	}

	for _, tag := range tags {
		if contains(availableTagsList, tag) {
			if !checkImageExists(dockerClient, tag) {
				fmt.Printf("Language %s has not been installed.\n", tag)
			} else {
				fmt.Printf("Updating %s.\n", tag)
				updateLanguage(dockerClient, tag)
			}
		} else {
			log.Fatalf("Language %s not a valid candidate", tag)
		}
	}
	return
}
