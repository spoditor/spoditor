package annotation

import (
	"regexp"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	Prefix    = "ssarg.io/"
	Separator = "_"
)

var log = logf.Log.WithName("annotations")

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

var _ QualifiedAnnotationCollector = CollectorFunc(nil)

var Collector QualifiedAnnotationCollector = defaultCollector

var defaultCollector CollectorFunc = func(accessor metav1.ObjectMetaAccessor) map[QualifiedName]string {
	m := map[QualifiedName]string{}
	for k, v := range accessor.GetObjectMeta().GetAnnotations() {
		ll := log.WithValues("key", k, "value", v)
		if strings.HasPrefix(k, Prefix) {
			ll.Info("found ss-arg annotation")
			n := strings.TrimPrefix(k, Prefix)
			i := strings.LastIndex(n, Separator)
			if i == -1 {
				ll.Info("dynamic argumentation")
				m[QualifiedName{Name: n}] = v
			} else {
				ll.Info("designated argumentation")
				m[QualifiedName{
					Qualifier: n[i+1:],
					Name:      n[:i],
				}] = v
			}
		} else {
			ll.Info("skip irrelevant annotation")
		}
	}
	return m
}

type PodQualifier func(int, string) bool

var CommonPodQualifier PodQualifier = func(ordinal int, q string) bool {
	ll := log.WithValues("ordinal", ordinal, "qualifier", q)
	if q == "" {
		ll.Info("pod is always included for dynamic argumentation")
		return true
	}
	bounds := strings.Split(q, "-")
	if b, err := regexp.MatchString("^\\d+-\\d+$", q); err == nil && b {
		min, _ := strconv.Atoi(bounds[0])
		max, _ := strconv.Atoi(bounds[1])
		ll.Info("check ordinal against range", "min", min, "max", max)
		return ordinal >= min && ordinal <= max
	}
	if b, err := regexp.MatchString("^\\d+$", q); err == nil && b {
		i, _ := strconv.Atoi(bounds[0])
		ll.Info("check ordinal against single exact number", "number", i)
		return ordinal == i
	}
	if b, err := regexp.MatchString("^\\d+-$", q); err == nil && b {
		min, _ := strconv.Atoi(bounds[0])
		ll.Info("check ordinal against lower bound", "min", min)
		return ordinal >= min
	}
	if b, err := regexp.MatchString("^-\\d+$", q); err == nil && b {
		max, _ := strconv.Atoi(bounds[1])
		ll.Info("check ordinal against upper bound", "max", max)
		return ordinal <= max
	}
	return false
}
