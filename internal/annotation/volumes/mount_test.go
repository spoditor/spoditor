package volumes

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/spoditor/spoditor/internal/annotation"
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
		want    *v1.PodSpec
		wantErr bool
	}{
		{
			name: "wrong config type",
			args: args{
				spec:    nil,
				ordinal: 0,
				cfg:     nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "do nothing because ordinal doesn't qualify",
			args: args{
				spec:    &v1.PodSpec{},
				ordinal: 0,
				cfg: &mountConfig{
					qualifier: "1-2",
					cfg:       nil,
				},
			},
			want:    &v1.PodSpec{},
			wantErr: false,
		},
		{
			name: "mount configmap as volume",
			args: args{
				spec: &v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "main-container",
						},
					},
				},
				ordinal: 0,
				cfg: &mountConfig{
					qualifier: "",
					cfg: &mountConfigValue{
						Volumes: []v1.Volume{
							{
								Name: "my-config",
								VolumeSource: v1.VolumeSource{
									ConfigMap: &v1.ConfigMapVolumeSource{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "my-configmap",
										},
									},
								},
							},
						},
						Containers: []v1.Container{
							{
								Name: "main-container",
								VolumeMounts: []v1.VolumeMount{
									{
										Name:      "my-config",
										MountPath: "/etc/my-config",
									},
								},
							},
						},
					},
				},
			},
			want: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "main-container",
						VolumeMounts: []v1.VolumeMount{
							{
								Name:      "my-config",
								MountPath: "/etc/my-config",
							},
						},
					},
				},
				Volumes: []v1.Volume{
					{
						Name: "my-config",
						VolumeSource: v1.VolumeSource{
							ConfigMap: &v1.ConfigMapVolumeSource{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "my-configmap-0",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "mount secret as volume",
			args: args{
				spec: &v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "main-container",
						},
					},
				},
				ordinal: 0,
				cfg: &mountConfig{
					qualifier: "",
					cfg: &mountConfigValue{
						Volumes: []v1.Volume{
							{
								Name: "my-secret",
								VolumeSource: v1.VolumeSource{
									Secret: &v1.SecretVolumeSource{
										SecretName: "my-secret",
									},
								},
							},
						},
						Containers: []v1.Container{
							{
								Name: "main-container",
								VolumeMounts: []v1.VolumeMount{
									{
										Name:      "my-secret",
										MountPath: "/etc/my-secret",
									},
								},
							},
						},
					},
				},
			},
			want: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "main-container",
						VolumeMounts: []v1.VolumeMount{
							{
								Name:      "my-secret",
								MountPath: "/etc/my-secret",
							},
						},
					},
				},
				Volumes: []v1.Volume{
					{
						Name: "my-secret",
						VolumeSource: v1.VolumeSource{
							Secret: &v1.SecretVolumeSource{
								SecretName: "my-secret-0",
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &MountHandler{}
			if err := h.Mutate(tt.args.spec, tt.args.ordinal, tt.args.cfg); (err != nil) != tt.wantErr {
				t.Errorf("Mutate() error = %v, wantErr %v", err, tt.wantErr)
			} else if !reflect.DeepEqual(tt.args.spec, tt.want) {
				t.Errorf("Mutate() = %v, want %v", tt.args.spec, tt.want)
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
					Name:      MountVolume,
				}: func() string {
					b, _ := json.Marshal(c.cfg)
					fmt.Println(string(b))
					return string(b)
				}(),
			}},
			want:    c,
			wantErr: false,
		},
		{
			name: "explicit json",
			p:    parser,
			args: args{annotations: map[annotation.QualifiedName]string{
				annotation.QualifiedName{
					Qualifier: "1-2",
					Name:      MountVolume,
				}: "{\"volumes\":[{\"name\": \"my-volume\", \"configMap\":{\"name\":\"my-configmap\"}}],\"containers\":[{\"name\":\"nginx\", \"volumeMounts\":[{\"name\":\"my-volume\",\"mountPath\":\"/etc/configmaps/my-volume\"}]}]}",
			}},
			want: &mountConfig{
				qualifier: "1-2",
				cfg: &mountConfigValue{
					Volumes: []v1.Volume{
						{
							Name: "my-volume",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "my-configmap",
									},
								},
							},
						},
					},
					Containers: []v1.Container{
						{
							Name: "nginx",
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "my-volume",
									MountPath: "/etc/configmaps/my-volume",
								},
							},
						},
					},
				},
			},
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
