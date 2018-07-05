package cmd

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/user"
	"time"

	rice "github.com/GeertJohan/go.rice"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	docker "github.com/docker/docker/client"
)

func checkImageExists(dockerClient *docker.Client, tag string) bool {
	if y, _, err := dockerClient.ImageInspectWithRaw(context.Background(), fmt.Sprintf("din/%s", tag)); err != nil {
		fmt.Println(err)
		return false
	} else {
		fmt.Println(y)
		return true
	}
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

	opts := types.ImageBuildOptions{
		Tags:       []string{fmt.Sprintf("din/%s", tag)},
		Dockerfile: fmt.Sprintf("%s.docker", tag),
	}
	_, err = dockerClient.ImageBuild(context.Background(), buffer, opts)
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
	opts := types.ImageBuildOptions{
		Tags:       []string{"din/base"},
		Dockerfile: "Dockerfile",
	}
	_, err = dockerClient.ImageBuild(context.Background(), buffer, opts)
	if err != nil {
		return err
	}
	return nil
}

func runImage(dockerClient *docker.Client, tag string, cmd string) error {
	currentUser, _ := user.Current()
	currentDir, _ := os.Getwd()

	environment := []string{
		fmt.Sprintf("DIN_ENV_PWD=%s", currentDir),
		fmt.Sprintf("DIN_ENV_UID=%s", currentUser.Uid),
		fmt.Sprintf("DIN_ENV_PWD=%s", currentUser.Username),
		fmt.Sprintf("DIN_COMMAND=%s", cmd),
	}

	createOpts := &container.Config{
		Env:   environment,
		Image: fmt.Sprintf("din/%s", tag),
	}

	container, err := dockerClient.ContainerCreate(context.Background(), createOpts, &container.HostConfig{}, &network.NetworkingConfig{}, fmt.Sprintf("din-%s", tag))
	if err != nil {
		return err
	}
	defer dockerClient.ContainerRemove(context.Background(), container.ID, types.ContainerRemoveOptions{Force: true})

	err = dockerClient.ContainerStart(context.Background(), container.ID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}
	timeout := 10 * time.Second
	defer dockerClient.ContainerStop(context.Background(), container.ID, &timeout)

	execConfig := types.ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          []string{cmd},
	}

	exec, err := dockerClient.ContainerExecCreate(context.Background(), container.ID, execConfig)
	if err != nil {
		return err
	}
	err = dockerClient.ContainerExecStart(context.Background(), exec.ID, types.ExecStartCheck{
		Detach: true,
		Tty:    true,
	})
	if err != nil {
		return err
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
