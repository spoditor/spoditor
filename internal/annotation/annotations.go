package annotation

import (
	"regexp"
	"strconv"
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

type PodQualifier func(int, string) bool

var CommonPodQualifier PodQualifier = func(ordinal int, q string) bool {
	if q == "" {
		return true
	}
	if b, err := regexp.MatchString("\\d-\\d", q); err != nil || !b {
		return false
	}
	bounds := strings.Split(q, "-")
	min, _ := strconv.Atoi(bounds[0])
	max, _ := strconv.Atoi(bounds[1])
	return ordinal >= min && ordinal <= max
}
