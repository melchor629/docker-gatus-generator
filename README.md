# docker-gatus-generator

Generate gatus config from containers using go templates. It is also possible to generate other kind of stuff using this as well. Just grab the image and have fun 😎

## Composerino

The image is located here: `melchor9000/docker-gatus-generator`.

```yaml
services:
  docker-gatus-generator:
    image: melchor9000/docker-gatus-generator
    container_name: docker-gatus-generator
    env:
      GATUS_FOLDER: /gatus/generated
      GATUS_TEMPLATE_TEST: |
        endpoints:
          {{- range $container := .Containers }}
          - name: {{ $container.Name }}
            group: test
            url: {{ $container | getUrlFromExposed }}
            interval: 60s
            conditions: ['[STATUS] == 200']
            {{- if $container.Gatus.Conditions }}
            conditions: {{ $container.Gatus.Conditions }}
            {{- end }}
          {{- end }}
    volumes:
      - gatus-config:/gatus
```

## How?

1. Add a list of containers using `gatus.enable=true` label
2. Each container assign at least one template with `gatus.template=test,...` label (multiple values allowed)
3. Optionally, add more [labels](#container-labels)
4. Configure the container with `GATUS_FOLDER` and `GATUS_TEMPLATES_TEST` (see [settings](#settings) for more)
5. Run the service
    1. The service looks for containers (running or not) that has the `gatus.enable=true` label
    2. Extracts some information including all [gatus labels](#container-labels)
    3. Reads each template from the `GATUS_TEMPLATE_FILE_{template}`/`GATUS_TEMPLATE_{template}`
    4. Optionally expands those templates with helper templates from `GATUS_TEMPLATES_PATH`
    5. Renders each template and puts the output of each in `GATUS_FOLDER`

## Settings

- `GATUS_FOLDER`: folder where the rendered templates will be put
- `GATUS_IP`: when using `getUrlFromExposed`, serves as default value for when the host IP cannot be detected
- `GATUS_TEMPLATES_PATH`: optional path where helper templates are located
- `GATUS_TEMPLATE_FILE_{template}`: for the specified `{template}`, reads the given file as template source
- `GATUS_TEMPLATE_{template}`: for the specified `{template}`, uses the contents of the variable as template source. this takes precedence over above variable
- `GATUS_LOG_LEVEL`: configures the log level (defaults to `info`)
- `DOCKER_HOST`: overrides the default host for the docker server
- `DOCKER_API_VERSION`: changes the API version to use (use at your own risk)
- `DOCKER_CERT_PATH`: specify the directory from which to load the TLS certificates
- `DOCKER_TLS_VERIFY`: enables or disables TLS verification (disabled by default)

## Template functions

Templates uses go [`text/template`](https://pkg.go.dev/text/template) for templates. Is expanded with 
[`sprig`](https://masterminds.github.io/sprig/) for additional utilities.

The input of the template has the following schema:

```go
. := TemplateInput{...}

type TemplateInput struct {
    Containers []Container // for non-go users that means Container[] or Array<Container>
    Hostname   string
    Ip         string
}

type Container struct {
    ID      string
    Name    string
    Summary container.Summary // https://pkg.go.dev/github.com/moby/moby/api/types/container#Summary
    Inspect container.InspectResponse // https://pkg.go.dev/github.com/moby/moby/api/types/container#InspectResponse
    Traefik *traefikDynamic.Configuration // https://pkg.go.dev/github.com/traefik/traefik/v3/pkg/config/dynamic#Configuration
    Gatus   *GatusConfig // for non-go users, ignore the pointer *
}

type GatusConfig struct {
    Path       string
    Scheme     string
    Port       string
    Conditions string
}
```

Additionaly, these functions can be used in the templates:

- `containerName`: extracts the container name from the container summary `{{ container.Summary | containerName }}`
- `getUrlFromDocker`: tries to generate an URL for the service using the IP and port from the docker network
- `getUrlFromExposed`: tries to generate an URL for the service using the host IP and the exposed port
- `getUrlFromTraefik`: tries to generate an URL from traefik labels
- `include`: `{{ template "name" }}` but with super-powers (can be used with pipelines or stored in variables), is like the helm one

## Container labels

There are a set of labels to customize the generated output in the templates.

- `gatus.enable`: if set to `true`, then the container will be considered to be used for the templates
- `gatus.template`: a comma-separated list of templates that the container will be sent to
- `gatus.path`: optional label to configure a path for the `getUrlFrom*` functions
- `gatus.port`: optional port to use for the `getUrlFrom*` functions, useful when there are mutliple exposed ports
- `gatus.scheme`: optional scheme for the service (defaults to `http`)
- `gatus.name`: sets a name for the container (defaults to the container name)
- `gatus.{template}.path`: same as `gatus.path` but just applies to the specified template
- `gatus.{template}.port`: same as `gatus.port` but just applies to the specified template
- `gatus.{template}.scheme`: same as `gatus.scheme` but just applies to the specified template
- `gatus.{template}.name`: same as `gatus.name` but just applies to the specified template
