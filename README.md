# Packer Dockerfile post-processor

DEPRECATED: This functionality has found it's way into the main packer binary. This repo is no longer maintained.

This is a [Packer](http://packer.io/) post-processor plugin which allows setting Docker metadata on an artifact from the [docker-tag](https://packer.io/docs/post-processors/docker-tag.html) post-processor.

Normally, Docker images built using Packer cannot include entrypoint, cmd, user, environment variables and other metadata that is available in Dockerfiles. This plugin will create a temporary Dockerfile and run `docker build` in an annonymous context. Most Dockerfile instructions are supported as json parametersp


# Caveats

* `RUN`, `ADD`, `COPY`, and `ONBUILD` are not supported because packer provisioners should be used for their functionality.
* Docker [no longer supports digests in `FROM`](https://github.com/docker/docker/commit/b8301005ffe66fb15a64735deeae707595543a92), thus a chained docker-tag post-processor is required.


# Usage

In your packer template, configure the post processor:

    {
      ...
      "post-processors": [
	    [
          {
            "type": "docker-tag",
            "repository": "localhost/example"
          },
          {
            "type": "docker-dockerfile",
            "volume": ["/data"]
            "expose": [8080],
            "entrypoint": ["/entrypoint.sh"],
            "cmd": ["bash"],
            "env": {
              "FOO": "bar"
            }
          }
        ]
      ]
      ...
    }

`cmd` and `entrypoint` can have either array or string values, this mirrors the Dockerfile format and functionality

See the [Dockerfile reference](http://docs.docker.com/reference/builder/) for details.



# Building

Install the necessary dependencies

    $ go get -d ./...

To compile the Packer plugin, run `go build`.

    $ go build

Put the binary `packer-post-processor-docker-dockerfile` into the `bin` directory of your choice


## Acknowledgement

* [Avishai Ish-Shalom's original plugin](https://github.com/avishai-ish-shalom/packer-post-processor-docker-dockerfile)
* [James G. Kim's re-write](https://github.com/jgkim/packer-post-processor-docker-dockerfile)


## Sponsors

This plugin was made possible by [Shiftgig](https://www.shiftgig.com/)


## License

This plugin is released under the Apache License, Version 2.0.


## Support

Please file an issue on the github repository if you think anything isn't working properly or an improvement is required.

This plugin was developed against Docker 1.8.1 and Packer 0.8.6.
