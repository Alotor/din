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
	"context"
	"fmt"
	"log"

	types "github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "deletes a language container",
	Long:  `deletes a language container`,
	Run:   cleanCmdF,
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}

func cleanCmdF(cmd *cobra.Command, tags []string) {
	dockerClient, err := docker.NewEnvClient()
	if err != nil {
		log.Fatalf("Unnable to connect to docker: %v", err)
	}

	// TODO: HANDLE THE ERROR
	availableTagsList, _ := availableTags()
	if len(tags) == 0 && confirm("delete all images") {
		for _, tag := range availableTagsList {
			if checkImageExists(dockerClient, tag) {
				tags = append(tags, tag)
			}
		}
	}

	for _, tag := range tags {
		if contains(availableTagsList, tag) {
			if !checkImageExists(dockerClient, tag) {
				fmt.Printf("There is nothing to clean for language %s.\n", tag)
			} else {
				fmt.Printf("Cleaning %s.\n", tag)
				opts := types.ImageRemoveOptions{Force: true}
				dockerClient.ImageRemove(context.Background(), fmt.Sprintf("din/%s", tag), opts)
			}
		} else {
			log.Fatalf("Language %s not found", tag)
		}
	}
	return
}
