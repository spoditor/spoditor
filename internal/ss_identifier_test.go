package internal

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLabelSSPodIdentifier_Extract(t *testing.T) {
	type args struct {
		accessor metav1.ObjectMetaAccessor
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   int
		wantErr bool
	}{
		{
			name: "missing label",
			args: args{
				accessor: &v1.Pod{},
			},
			want:    "",
			want1:   -1,
			wantErr: true,
		},
		{
			name: "unexpected label format",
			args: args{
				accessor: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"statefulset.kubernetes.io/pod-name": "wrong-label"},
					},
				},
			},
			want:    "",
			want1:   -1,
			wantErr: true,
		},
		{
			name: "success",
			args: args{
				accessor: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"statefulset.kubernetes.io/pod-name": "dummy-ss-0"},
					},
				},
			},
			want:    "dummy-ss",
			want1:   0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := LabelSSPodIdentifier
			got, got1, err := d.Extract(tt.args.accessor)
			if (err != nil) != tt.wantErr {
				t.Errorf("Extract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Extract() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Extract() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
