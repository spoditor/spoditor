package annotation

import (
	"reflect"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCollectorFunc_Collect(t *testing.T) {
	type args struct {
		accessor v1.ObjectMetaAccessor
	}
	tests := []struct {
		name string
		c    QualifiedAnnotationCollector
		args args
		want map[QualifiedName]string
	}{
		{
			name: "success",
			c:    Collector,
			args: args{accessor: &v1.ObjectMeta{
				Annotations: map[string]string{
					"ssarg.io/mount-configmap":     "dummy value",
					"ssarg.io/mount-configmap_1-2": "dummy value",
				},
			}},
			want: map[QualifiedName]string{
				QualifiedName{
					Name: "mount-configmap",
				}: "dummy value",
				QualifiedName{
					Qualifier: "1-2",
					Name:      "mount-configmap",
				}: "dummy value",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Collect(tt.args.accessor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Collect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPodQualifier(t *testing.T) {
	type args struct {
		ordinal int
		q       string
	}
	tests := []struct {
		name string
		q    PodQualifier
		args args
		want bool
	}{
		{
			name: "apply all qualifier",
			q:    CommonPodQualifier,
			args: args{
				ordinal: 0,
			},
			want: true,
		},
		{
			name: "unexpected qualifier",
			q:    CommonPodQualifier,
			args: args{
				ordinal: 0,
				q:       "not even a qualifier",
			},
			want: false,
		},
		{
			name: "common pod qualifier excluded from a range",
			q:    CommonPodQualifier,
			args: args{
				ordinal: 0,
				q:       "1-2",
			},
			want: false,
		},
		{
			name: "common pod qualifier included in a range",
			q:    CommonPodQualifier,
			args: args{
				ordinal: 0,
				q:       "00-20",
			},
			want: true,
		},
		{
			name: "common pod qualifier exact single pod",
			q:    CommonPodQualifier,
			args: args{
				ordinal: 0,
				q:       "000",
			},
			want: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.q(test.args.ordinal, test.args.q); got != test.want {
				t.Errorf("PodQualifier() = %v, want %v", got, test.want)
			}
		})
	}
}
