package internal

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var podIdentifierLog = logf.Log.WithName("pod_identifier")

type SSPodIdentifier interface {
	Extract(accessor v1.ObjectMetaAccessor) (string, int, error)
}

type SSPodIdentifierFunc func(v1.ObjectMetaAccessor) (string, int, error)

func (f SSPodIdentifierFunc) Extract(accessor v1.ObjectMetaAccessor) (string, int, error) {
	return f(accessor)
}

type LabelSSPodIdentifier struct{}

func (d *LabelSSPodIdentifier) Extract(accessor v1.ObjectMetaAccessor) (string, int, error) {
	l, ok := accessor.GetObjectMeta().GetLabels()["statefulset.kubernetes.io/pod-name"]
	if !ok {
		podIdentifierLog.Info("statefulset label not found", "label", "statefulset.kubernetes.io/pod-name")
		return "", -1, errors.New("missing statefulset label")
	}
	podIdentifierLog.Info("stateful pod name", "name", l)
	if b, err := regexp.MatchString(".+-\\d+", l); err != nil || !b {
		return "", -1, errors.New("unexpected label value")
	}
	i := strings.LastIndex(l, "-")
	ordinal, _ := strconv.Atoi(l[i+1:])
	return l[:i], ordinal, nil
}
