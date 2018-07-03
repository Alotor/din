package cmd

import (
	"archive/tar"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/user"

	rice "github.com/GeertJohan/go.rice"
	docker "github.com/fsouza/go-dockerclient"
)

func checkImageExists(dockerClient *docker.Client, tag string) bool {
	if _, err := dockerClient.InspectImage(fmt.Sprintf("din/%s", tag)); err != nil {
		return false
	}
	return true
}

func buildImage(dockerClient *docker.Client, tag string) error {
	box := rice.MustFindBox("../lang")

	dockerfile, err := box.Open(fmt.Sprintf("%s.docker", tag))
	if err != nil {
		return err
	}

	buffer, err := toTarBuffer(dockerfile)
	if err != nil {
		return err
	}

	opts := docker.BuildImageOptions{
		Name:         fmt.Sprintf("din/%s", tag),
		Dockerfile:   fmt.Sprintf("%s.docker", tag),
		InputStream:  buffer,
		OutputStream: os.Stdout,
	}
	err = dockerClient.BuildImage(opts)
	if err != nil {
		return err
	}
	return nil
}

func buildBaseImage(dockerClient *docker.Client) error {
	box := rice.MustFindBox("../base")
	dockerfile, err := box.Open("Dockerfile")
	if err != nil {
		return err
	}
	dinscript, err := box.Open("din")
	if err != nil {
		return err
	}
	buffer, err := toTarBuffer(dockerfile, dinscript)
	if err != nil {
		return err
	}
	opts := docker.BuildImageOptions{
		Name:         "din/base",
		InputStream:  buffer,
		OutputStream: os.Stdout,
	}
	err = dockerClient.BuildImage(opts)
	if err != nil {
		return err
	}
	return nil
}

func runImage(tag string, cmd string) error {
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
