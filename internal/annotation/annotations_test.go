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
		c    CollectorFunc
		args args
		want map[QualifiedName]string
	}{
		{
			name: "success",
			c:    NewCollector(),
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
