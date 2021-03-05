package annotation

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Handler interface {
	Mutate(spec *corev1.PodSpec, ordinal int, cfg interface{}) error
	NewParser() Parser
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

var defaultCollector CollectorFunc = func(accessor metav1.ObjectMetaAccessor) map[QualifiedName]string {
	m := map[QualifiedName]string{}
	for k, v := range accessor.GetObjectMeta().GetAnnotations() {
		if strings.HasPrefix(k, "ssarg.io") {
			n := strings.TrimPrefix(k, "ssarg.io/")
			i := strings.LastIndex(n, "_")
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

func NewCollector() CollectorFunc {
	return defaultCollector
}
