package env

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spoditor/spoditor/internal/annotation"
	corev1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	MountEnv = "mount-env"
)

var log = logf.Log.WithName("mount_env")

type mountEnv struct {
	qualifier string
	cfg       *mountConfigValue
}

type mountConfigValue struct {
	EnvFrom    []corev1.EnvFromSource `json:"envFrom"`
	Containers []corev1.Container     `json:"containers"`
}

type MountHandler struct {
}

func (h *MountHandler) Mutate(spec *corev1.PodSpec, ordinal int, cfg interface{}) error {
	ll := log.WithValues("ordinal", ordinal)

	m, ok := cfg.(*mountEnv)
	if !ok {
		return fmt.Errorf("unexpected config type %T", m)
	}
	if should(ordinal, m.qualifier) {
		for _, v := range m.cfg.EnvFrom {
			if v.ConfigMapRef != nil {
				ll.Info("overwrite configmap name",
					"origin", v.ConfigMapRef.LocalObjectReference.Name,
					"new", v.ConfigMapRef.LocalObjectReference.Name+"-"+strconv.Itoa(ordinal))
				v.ConfigMapRef.LocalObjectReference.Name += "-" + strconv.Itoa(ordinal)
			}
			if v.SecretRef != nil {
				ll.Info("overwrite secret name",
					"origin", v.SecretRef.LocalObjectReference.Name,
					"new", v.SecretRef.LocalObjectReference.Name+"-"+strconv.Itoa(ordinal))
				v.SecretRef.LocalObjectReference.Name += "-" + strconv.Itoa(ordinal)
			}
		}
		for _, source := range m.cfg.Containers {
			for i := 0; i < len(spec.Containers); i++ {
				if source.Name == spec.Containers[i].Name {
					ll.Info("Add EnvFrom to container", "container", source.Name)
					spec.Containers[i].EnvFrom = append(spec.Containers[i].EnvFrom, m.cfg.EnvFrom...)
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

var parser annotation.ParserFunc = func(annotations map[annotation.QualifiedName]string) (interface{}, error) {
	for k, v := range annotations {
		ll := log.WithValues("qualifiedName", k, "value", v)
		if k.Name == MountEnv {
			ll.Info("parse config for mounting volumes")
			c := &mountConfigValue{}
			if err := json.Unmarshal([]byte(v), c); err == nil {
				return &mountEnv{
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
