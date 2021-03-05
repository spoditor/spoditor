package configmap

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/simingweng/ss-argumentor/internal/annotation"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
)

func TestMountHandler_Mutate(t *testing.T) {
	type args struct {
		spec    *v1.PodSpec
		ordinal int
		cfg     interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &MountHandler{}
			if err := h.Mutate(tt.args.spec, tt.args.ordinal, tt.args.cfg); (err != nil) != tt.wantErr {
				t.Errorf("Mutate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_parserFunc_Parse(t *testing.T) {
	type args struct {
		annotations map[annotation.QualifiedName]string
	}
	c := &mountConfig{
		qualifier: "1-2",
		cfg: &mountConfigValue{
			Volumes: []v1.Volume{
				{
					Name: "dummy-vol",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "dummy-configmap",
							},
						},
					},
				},
			},
			Containers: []v1.Container{
				{
					Name: "dummy-container",
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "dummy-vol",
							MountPath: "/etc/configmaps/dummy",
						},
					},
				},
			},
		},
	}
	tests := []struct {
		name    string
		p       parserFunc
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name:    "no expected annotation",
			p:       parser,
			args:    args{annotations: map[annotation.QualifiedName]string{}},
			want:    nil,
			wantErr: false,
		},
		{
			name: "success",
			p:    parser,
			args: args{annotations: map[annotation.QualifiedName]string{
				annotation.QualifiedName{
					Qualifier: "1-2",
					Name:      MountConfigMaps,
				}: func() string {
					b, _ := json.Marshal(c.cfg)
					fmt.Println(string(b))
					return string(b)
				}(),
			}},
			want:    c,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.p.Parse(tt.args.annotations)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() got = %v, want %v", got, tt.want)
			}
		})
	}
}
