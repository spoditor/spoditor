package volumes

import (
	"fmt"
	"strconv"

	"github.com/spoditor/spoditor/internal/annotation"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	MountVolume = "mount-volume"
)

var log = logf.Log.WithName("mount_volume")

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

var _ annotation.Parser = parserFunc(nil)

type MountHandler struct {
}

func (h *MountHandler) Mutate(spec *corev1.PodSpec, ordinal int, cfg interface{}) error {
	ll := log.WithValues("ordinal", ordinal)
	m, ok := cfg.(*mountConfig)
	if !ok {
		return fmt.Errorf("unexpected config type %T", m)
	}
	if should(ordinal, m.qualifier) {
		ll.Info("pod should be applicable")
		for _, v := range m.cfg.Volumes {
			if v.ConfigMap != nil {
				ll.Info("overwrite configmap name",
					"volume", v.Name,
					"origin", v.ConfigMap.LocalObjectReference.Name,
					"new", v.ConfigMap.LocalObjectReference.Name+"-"+strconv.Itoa(ordinal))
				v.ConfigMap.LocalObjectReference.Name += "-" + strconv.Itoa(ordinal)
			}
			if v.Secret != nil {
				ll.Info("overwrite secret name",
					"volume", v.Name,
					"origin", v.Secret.SecretName,
					"new", v.Secret.SecretName+"-"+strconv.Itoa(ordinal))
				v.Secret.SecretName += "-" + strconv.Itoa(ordinal)
			}
		}
		spec.Volumes = append(spec.Volumes, m.cfg.Volumes...)
		for _, source := range m.cfg.Containers {
			for i := 0; i < len(spec.Containers); i++ {
				if source.Name == spec.Containers[i].Name {
					ll.Info("mount volumes to container", "container", source.Name)
					spec.Containers[i].VolumeMounts = append(spec.Containers[i].VolumeMounts, source.VolumeMounts...)
				}
			}
		}
	} else {
		log.Info("qualifier excludes this pod")
	}
	return nil
}

func (h *MountHandler) GetParser() annotation.Parser {
	return parser
}

var _ annotation.Handler = &MountHandler{}

var parser parserFunc = func(annotations map[annotation.QualifiedName]string) (interface{}, error) {
	for k, v := range annotations {
		ll := log.WithValues("qualifiedName", k, "value", v)
		if k.Name == MountVolume {
			ll.Info("parse config for mounting volumes")
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
