package internal

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	StatefulSetNameKey = "Key_StatefulSet_Name"
	PodOrdinalKey      = "Key_Pod_Ordinal"
)

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
		return "", -1, errors.New("missing statefulset label")
	}
	if b, err := regexp.MatchString(".+-\\d+", l); err != nil || !b {
		return "", -1, errors.New("unexpected label value")
	}
	i := strings.LastIndex(l, "-")
	ordinal, _ := strconv.Atoi(l[i+1:])
	return l[:i], ordinal, nil
}
