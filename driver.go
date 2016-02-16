package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"

	"github.com/mitchellh/packer/builder/docker"
)

type Driver interface {
	docker.Driver

	BuildImage(*bytes.Buffer) (string, error)
}

type DockerDriver struct {
	*docker.DockerDriver
}

// Build an image from a compiled Dockerfile
// Return the ID for the new image
func (d *DockerDriver) BuildImage(df *bytes.Buffer) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	c := exec.Command("docker", "build", "--force-rm=true", "--no-cache=true", "--quiet=true", "-")
	c.Stdin = df
	c.Stdout = &stdout
	c.Stderr = &stderr

	if err := c.Start(); err != nil {
		return "", err
	}

	if err := c.Wait(); err != nil {
		err = fmt.Errorf("Error building image: %s\nStderr: %s", err, stderr.String())
		return "", err
	}

	p := regexp.MustCompile("sha256:([a-f0-9]+)")
	m := p.FindStringSubmatch(stdout.String())
	if m == nil {
		err := fmt.Errorf("Error parsing `docker build` output: %s", stdout.String())

		return "", err
	}
	id := m[len(m)-1]

	return id, nil
}
