package volumes

import (
	"fmt"
	"strconv"

	"github.com/simingweng/ss-argumentor/internal/annotation"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	MountVolume = "mount-volume"
)

var mountCfgMapLog = logf.Log.WithName("mount_volume")

type mountConfig struct {
	qualifier string
	cfg       *mountConfigValue
}

type mountConfigValue struct {
	Volumes    []corev1.Volume    `json:"volumes"`
	Containers []corev1.Container `json:"containers"`
}

type parserFunc func(map[annotation.QualifiedName]string) (interface{}, error)

func (p parserFunc) Parse(annotations map[annotation.QualifiedName]string) (interface{}, error) {
	return p(annotations)
}

type MountHandler struct {
}

func (h *MountHandler) Mutate(spec *corev1.PodSpec, ordinal int, cfg interface{}) error {
	m, ok := cfg.(*mountConfig)
	if !ok {
		return fmt.Errorf("unexpected config type %T", m)
	}
	if should(ordinal, m.qualifier) {
		for _, v := range m.cfg.Volumes {
			if v.ConfigMap != nil {
				v.ConfigMap.LocalObjectReference.Name += "-" + strconv.Itoa(ordinal)
			}
			if v.Secret != nil {
				v.Secret.SecretName += "-" + strconv.Itoa(ordinal)
			}
		}
		spec.Volumes = append(spec.Volumes, m.cfg.Volumes...)
		for _, source := range m.cfg.Containers {
			for i := 0; i < len(spec.Containers); i++ {
				if source.Name == spec.Containers[i].Name {
					spec.Containers[i].VolumeMounts = append(spec.Containers[i].VolumeMounts, source.VolumeMounts...)
				}
			}
		}
	} else {
		mountCfgMapLog.Info("qualifier excludes this pod")
	}
	return nil
}

func (h *MountHandler) GetParser() annotation.Parser {
	return parser
}

var parser parserFunc = func(annotations map[annotation.QualifiedName]string) (interface{}, error) {
	for k, v := range annotations {
		if k.Name == MountVolume {
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

var should = annotation.CommonPodQualifier
