package i18n

import (
	"github.com/lazygophers/utils/json"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
	"strings"
)

type Localizer interface {
	Unmarshal(body []byte, v any) (err error)
}

var localizer = map[string]Localizer{
	"json": jsonLocalizer,
	"yaml": yamlLocalizer,
	"yml":  yamlLocalizer,
	"toml": tomlLocalizer,
}

var (
	jsonLocalizer = NewLocalizerHandle(json.Unmarshal)

	yamlLocalizer = NewLocalizerHandle(yaml.Unmarshal)

	tomlLocalizer = NewLocalizerHandle(toml.Unmarshal)
)

func RegisterLocalizer(name string, v Localizer) {
	localizer[name] = v
}

func GetLocalizer(name string) (Localizer, bool) {
	if strings.HasPrefix(name, ".") {
		name = name[1:]
	}

	l, ok := localizer[name]
	return l, ok
}

var _ Localizer = (*LocalizerHandle)(nil)

type LocalizerHandle struct {
	unmarshal func(body []byte, v any) (err error)
}

func (p *LocalizerHandle) Unmarshal(body []byte, v any) (err error) {
	return p.unmarshal(body, v)
}

func NewLocalizerHandle(unmarshal func(body []byte, v any) (err error)) *LocalizerHandle {
	return &LocalizerHandle{
		unmarshal: unmarshal,
	}
}
