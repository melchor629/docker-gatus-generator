package dockergatusgenerator

import (
	"os"
	"slices"
	"strings"
)

type TemplateInputs map[string]TemplateInput

func (l TemplateInputs) Append(templateName string, containerSummary Container) {
	list := l[templateName]
	list.Containers = append(list.Containers, containerSummary)
	l[templateName] = list
}

func (l TemplateInputs) Finish(config Config) TemplateInputs {
	for pos, i := range l {
		slices.SortFunc(
			i.Containers,
			func(a Container, b Container) int {
				return strings.Compare(a.Name, b.Name)
			},
		)
		if i.Hostname = config.GetHostname(); i.Hostname == "" {
			i.Hostname, _ = os.Hostname()
		}
		i.Ip = config.GetIp()
		l[pos] = i
	}
	return l
}
