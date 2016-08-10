package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"text/template"

	"github.com/mitchellh/packer/builder/docker"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/post-processor/docker-tag"
	"github.com/mitchellh/packer/template/interpolate"
)

const BuilderId = "packer.post-processor.docker-dockerfile"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	From       string
	Maintainer string            `mapstructure:"maintainer"`
	Cmd        interface{}       `mapstructure:"cmd"`
	Label      map[string]string `mapstructure:"label"`
	Expose     []string          `mapstructure:"expose"`
	Env        map[string]string `mapstructure:"env"`
	Entrypoint interface{}       `mapstructure:"entrypoint"`
	Volume     []string          `mapstructure:"volume"`
	User       string            `mapstructure:"user"`
	WorkDir    string            `mapstructure:"workdir"`

	ctx interpolate.Context
}

type PostProcessor struct {
	Driver Driver

	config Config
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	return nil
}

func (p *PostProcessor) processVar(v interface{}) (string, error) {
	switch t := v.(type) {
	case []string:
		a := make([]string, 0, len(t))
		for _, item := range t {
			a = append(a, item)
		}
		r, _ := json.Marshal(a)
		return string(r), nil
	case []interface{}:
		a := make([]string, 0, len(t))
		for _, item := range t {
			a = append(a, item.(string))
		}
		r, _ := json.Marshal(a)
		return string(r), nil
	case string:
		return t, nil
	case nil:
		return "", nil
	}

	return "", fmt.Errorf("Unsupported variable type: %s", reflect.TypeOf(v))
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	if artifact.BuilderId() != dockertag.BuilderId {
		err := fmt.Errorf(
			"Unknown artifact type: %s\nCan only build with Dockerfile from Docker tag artifacts.",
			artifact.BuilderId())
		return nil, false, err
	}

	p.config.From = artifact.Id()

	template_str := `FROM {{ .From }}
{{ if .Maintainer }}MAINTAINER {{ .Maintainer }}
{{ end }}{{ if .Cmd }}CMD {{ process .Cmd }}
{{ end }}{{ if .Label }}{{ range $k, $v := .Label }}LABEL "{{ $k }}"="{{ $v }}"
{{ end }}{{ end }}{{ if .Expose }}EXPOSE {{ join .Expose " " }}
{{ end }}{{ if .Env }}{{ range $k, $v := .Env }}ENV {{ $k }} {{ $v }}
{{ end }}{{ end }}{{ if .Entrypoint }}ENTRYPOINT {{ process .Entrypoint }}
{{ end }}{{ if .Volume }}VOLUME {{ process .Volume }}
{{ end }}{{ if .User }}USER {{ .User }}
{{ end }}{{ if .WorkDir }}WORKDIR {{ .WorkDir }}{{ end }}`

	dockerfile := new(bytes.Buffer)
	template_writer := bufio.NewWriter(dockerfile)

	tmpl, err := template.New("Dockerfile").Funcs(template.FuncMap{
		"process": p.processVar,
		"join":    strings.Join,
	}).Parse(template_str)
	if err != nil {
		return nil, false, err
	}
	err = tmpl.Execute(template_writer, p.config)
	if err != nil {
		return nil, false, err
	}
	template_writer.Flush()
	log.Printf("Dockerfile:\n%s", dockerfile.String())

	driver := p.Driver
	if driver == nil {
		// If no driver is set, then we use the real driver
		driver = &DockerDriver{&docker.DockerDriver{Ctx: &p.config.ctx, Ui: ui}}
	}

	ui.Message("Building image from Dockerfile")
	id, err := driver.BuildImage(dockerfile)
	if err != nil {
		return nil, false, err
	}

	ui.Message("Destroying previously tagged image: " + p.config.From)
	err = artifact.Destroy()
	if err != nil {
		return nil, false, err
	}

	ui.Message("Tagging new image, " + id + ", as " + p.config.From)
	err = driver.TagImage(id, p.config.From, false)
	if err != nil {
		return nil, false, err
	}

	// Build the artifact
	artifact = &docker.ImportArtifact{
		BuilderIdValue: dockertag.BuilderId,
		Driver:         driver,
		IdValue:        p.config.From,
	}

	return artifact, true, nil
}
