package cmd

import (
	"fmt"
	"os"
	"strings"

	rice "github.com/GeertJohan/go.rice"
	docker "github.com/docker/docker/client"
)

func availableTags() ([]string, error) {
	box := rice.MustFindBox("../lang")
	tags := []string{}
	err := box.Walk("", func(path string, info os.FileInfo, err error) error {
		tags = append(tags, strings.Replace(info.Name(), ".docker", "", -1))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func executeDin(dockerClient *docker.Client, tag string, cmd string) error {
	if cmd == "" {
		cmd = tag
	}

	tags, err := availableTags()
	if err != nil {
		return err
	}

	if !contains(tags, tag) {
		return fmt.Errorf("ERROR: Can't find language: %s\n", tag)
	}

	if !checkImageExists(dockerClient, "base") {
		if err := buildBaseImage(dockerClient); err != nil {
			return fmt.Errorf("Cannot create the 'din/base' image")
		}
	}

	if !checkImageExists(dockerClient, tag) {
		if err := buildImage(dockerClient, tag); err != nil {
			return fmt.Errorf("Cannot create the 'din/%s' image", tag)
		}
	}

	return runImage(dockerClient, tag, cmd)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if strings.Compare(a, e) == 0 {
			return true
		}
	}
	return false
}

func installLanguage(dockerClient *docker.Client, tag string) error {
	if !checkImageExists(dockerClient, "base") {
		if err := buildBaseImage(dockerClient); err != nil {
			return fmt.Errorf("Cannot create the 'din/base' image.")
		}
	}

	if !checkImageExists(dockerClient, tag) {
		if err := buildImage(dockerClient, tag); err != nil {
			return fmt.Errorf("Cannot create the 'din/%s' image", tag)
		}
	}

	return nil
}

func updateLanguage(dockerClient *docker.Client, tag string) error {
	if !checkImageExists(dockerClient, tag) {
		return fmt.Errorf("Image 'din/%s' doesn't exist", tag)
	}

	if err := buildImage(dockerClient, tag); err != nil {
		return fmt.Errorf("Cannot update the 'din/%s' image", tag)
	}

	return nil
}

func confirm(action string) bool {
	var confirm string
	fmt.Printf("I'm about to proceed and %s.\n", action)
	fmt.Print("Are you sure? (y/n) ")
	fmt.Scanf("%s", &confirm)
	if confirm == "y" || confirm == "Y" {
		return true
	}
	return false
}
