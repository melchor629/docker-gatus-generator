endpoints:
{{- range $container := .Containers }}
- name: {{ $container.Name }}
  group: traefik
  url: {{ $container | getUrlFromTraefik }}
- name: {{ $container.Summary | containerName }}
  group: docker
  url: {{ $container | getUrlFromDocker }}
- name: {{ $container.Name }}
  group: exposed
  url: {{ $container | getUrlFromExposed }}
{{- end }}