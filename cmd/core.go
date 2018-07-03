package cmd

import (
	"archive/tar"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	rice "github.com/GeertJohan/go.rice"
	docker "github.com/fsouza/go-dockerclient"
)

func showCandidates() error {
	tags, err := availableTags()
	if err != nil {
		return err
	}
	for _, tag := range tags {
		fmt.Println(tag)
	}
	return nil
}

func toTarBuffer(inputs ...*rice.File) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for _, input := range inputs {
		stat, err := input.Stat()
		if err != nil {
			return nil, err
		}

		hdr := &tar.Header{Name: stat.Name(), Mode: int64(stat.Mode()), Size: stat.Size()}

		if err := tw.WriteHeader(hdr); err != nil {
			return nil, err
		}

		inputbuf := make([]byte, stat.Size())
		if _, err := input.Read(inputbuf); err != nil {
			return nil, err
		}
		if _, err := tw.Write(inputbuf); err != nil {
			return nil, err
		}
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}
	return &buf, nil
}

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

// TODO: RETURN THE POSSIBLE ERROR
func executeDin(dockerClient *docker.Client, tag string, cmd string) bool {
	if cmd == "" {
		cmd = tag
	}

	// TODO: HANDLE THE ERROR
	tags, _ := availableTags()

	if !contains(tags, tag) {
		fmt.Printf("ERROR: Can't find language: %s\n", tag)
		showCandidates()
		return false
	}

	if !checkImageExists(dockerClient, "base") {
		fmt.Println("Creating base image....")
		if !buildBaseImage(dockerClient) {
			fmt.Println("Cannot create the 'din/base' image. Exiting")
			return false
		}
		fmt.Println("Base image created.")
	}

	if !checkImageExists(dockerClient, tag) {
		if !buildImage(dockerClient, tag) {
			fmt.Printf("Cannot create the 'din/%s' image.\n", tag)
			return false
		}
	}

	return runImage(tag, cmd)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if strings.Compare(a, e) == 0 {
			return true
		}
	}
	return false
}

func runImage(tag string, cmd string) bool {
	currentUser, _ := user.Current()
	currentDir, _ := os.Getwd()

	dockerCmd := exec.Command(
		"docker", "run",
		"-it",
		"-e", fmt.Sprintf("DIN_ENV_PWD=\"%s\"", currentDir),
		"-e", fmt.Sprintf("DIN_ENV_UID=%s", currentUser.Uid),
		"-e", fmt.Sprintf("DIN_ENV_USER=%s", currentUser.Username),
		"-e", fmt.Sprintf("DIN_COMMAND=%s", cmd),
		"-v", fmt.Sprintf("%s:/home/%s", currentUser.HomeDir, currentUser.Username),
		fmt.Sprintf("din/%s", tag),
	)
	dockerCmd.Stdin = os.Stdin
	dockerCmd.Stdout = os.Stdout
	dockerCmd.Stderr = os.Stderr
	dockerCmd.Run()
	return true
}

func buildImage(dockerClient *docker.Client, tag string) bool {
	box := rice.MustFindBox("../lang")
	// TODO: HANDLE THE ERROR PROPERLY (PROPAGATING IT)
	dockerfile, err := box.Open(fmt.Sprintf("%s.docker", tag))
	if err != nil {
		return false
	}

	buffer, err := toTarBuffer(dockerfile)
	if err != nil {
		return false
	}

	opts := docker.BuildImageOptions{
		Name:         fmt.Sprintf("din/%s", tag),
		Dockerfile:   fmt.Sprintf("%s.docker", tag),
		InputStream:  buffer,
		OutputStream: os.Stdout,
	}
	err = dockerClient.BuildImage(opts)
	return err == nil
}

func buildBaseImage(dockerClient *docker.Client) bool {
	box := rice.MustFindBox("../base")
	dockerfile, err := box.Open("Dockerfile")
	if err != nil {
		log.Fatal(err.Error())
		return false
	}
	dinscript, err := box.Open("din")
	if err != nil {
		log.Fatal(err.Error())
		return false
	}
	buffer, err := toTarBuffer(dockerfile, dinscript)
	if err != nil {
		log.Fatal(err.Error())
		return false
	}
	opts := docker.BuildImageOptions{
		Name:         "din/base",
		InputStream:  buffer,
		OutputStream: os.Stdout,
	}
	err = dockerClient.BuildImage(opts)
	if err != nil {
		log.Fatal(err.Error())
	}
	return err == nil
}

func installLanguage(dockerClient *docker.Client, tag string) bool {
	if !checkImageExists(dockerClient, "base") {
		fmt.Println("Creating base image....")
		if !buildBaseImage(dockerClient) {
			fmt.Println("Cannot create the 'din/base' image. Exiting")
			return false
		}
		fmt.Println("Base image created.")
	}

	fmt.Println("Installing language {}...", tag)
	if !checkImageExists(dockerClient, tag) {
		if !buildImage(dockerClient, tag) {
			fmt.Printf("Cannot create the 'din/%s' image.\n", tag)
			return false
		}
	}
	fmt.Printf("Installed '%s'\n", tag)
	return true
}

func updateLanguage(dockerClient *docker.Client, tag string) bool {
	if !checkImageExists(dockerClient, tag) {
		fmt.Printf("The language %s has not been installed. Cancel update\n", tag)
		return false
	}

	fmt.Printf("Updating language %s...\n", tag)

	if !buildImage(dockerClient, tag) {
		fmt.Printf("Cannot update the 'din/%s' image.\n", tag)
		return false
	}

	fmt.Printf("Updated '%s'\n", tag)
	return true
}

func getScriptPath() string {
	_, filename, _, _ := runtime.Caller(1)
	dir, _ := filepath.Abs(filepath.Dir(filename))
	return dir
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

func checkImageExists(dockerClient *docker.Client, tag string) bool {
	_, err := dockerClient.InspectImage(fmt.Sprintf("din/%s", tag))
	return err == nil
}
