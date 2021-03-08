package annotation

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Prefix    = "ssarg.io/"
	Separator = "_"
)

type Handler interface {
	Mutate(spec *corev1.PodSpec, ordinal int, cfg interface{}) error
	GetParser() Parser
}

type Parser interface {
	Parse(annotations map[QualifiedName]string) (interface{}, error)
}

type QualifiedName struct {
	Qualifier string
	Name      string
}

type QualifiedAnnotationCollector interface {
	Collect(accessor metav1.ObjectMetaAccessor) map[QualifiedName]string
}

type CollectorFunc func(metav1.ObjectMetaAccessor) map[QualifiedName]string

func (c CollectorFunc) Collect(accessor metav1.ObjectMetaAccessor) map[QualifiedName]string {
	return c(accessor)
}

var Collector QualifiedAnnotationCollector = defaultCollector

var defaultCollector CollectorFunc = func(accessor metav1.ObjectMetaAccessor) map[QualifiedName]string {
	m := map[QualifiedName]string{}
	for k, v := range accessor.GetObjectMeta().GetAnnotations() {
		if strings.HasPrefix(k, Prefix) {
			n := strings.TrimPrefix(k, Prefix)
			i := strings.LastIndex(n, Separator)
			if i == -1 {
				m[QualifiedName{Name: n}] = v
			} else {
				m[QualifiedName{
					Qualifier: n[i+1:],
					Name:      n[:i],
				}] = v
			}
		}
	}
	return m
}
