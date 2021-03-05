package configmap

import (
	"github.com/simingweng/ss-argumentor/internal/annotation"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
)

const (
	MountConfigMaps = "mount-configmap"
)

type mountConfig struct {
	qualifier string
	cfg       *mountConfigValue
}

type mountConfigValue struct {
	Volumes    []corev1.Volume
	Containers []corev1.Container
}

type parserFunc func(map[annotation.QualifiedName]string) (interface{}, error)

func (p parserFunc) Parse(annotations map[annotation.QualifiedName]string) (interface{}, error) {
	return p(annotations)
}

type MountHandler struct {
}

func (h *MountHandler) Mutate(spec *corev1.PodSpec, ordinal int, cfg interface{}) error {
	_ = cfg.(*mountConfig)
	return nil
}

func (h *MountHandler) NewParser() annotation.Parser {
	return parser
}

var parser parserFunc = func(annotations map[annotation.QualifiedName]string) (interface{}, error) {
	for k, v := range annotations {
		if k.Name == MountConfigMaps {
			c := &mountConfigValue{}
			if err := json.Unmarshal([]byte(v), c); err == nil {
				return &mountConfig{
					qualifier: k.Qualifier,
					cfg:       c,
				}, nil
			} else {
				return nil, err
			}
		}
	}
	return nil, nil
}
