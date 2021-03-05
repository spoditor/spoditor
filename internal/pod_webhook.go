package internal

import (
	"context"
	"net/http"

	"github.com/simingweng/ss-argumentor/internal/annotation"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=ignore,sideEffects=None,groups="",resources=pods,verbs=create;update,versions=v1,name=mpod.ssarg.io,admissionReviewVersions={v1,v1beta1}

// log is for logging in this package.
var argumentorlog = logf.Log.WithName("pod-argumentor")

type PodArgumentor struct {
	decoder   *admission.Decoder
	SSPodId   SSPodIdentifier
	handlers  []annotation.Handler
	Collector annotation.QualifiedAnnotationCollector
}

func (r *PodArgumentor) Handle(c context.Context, request admission.Request) admission.Response {
	pod := &v1.Pod{}
	err := r.decoder.Decode(request, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// mutate the fields in pod
	argumentorlog.Info("received request for pod", "pod", pod)
	_, ordinal, err := r.SSPodId.Extract(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	for _, h := range r.handlers {
		c, err := h.NewParser().Parse(r.Collector.Collect(pod))
		if err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
		err = h.Mutate(&pod.Spec, ordinal, c)
		if err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
	}

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(request.Object.Raw, marshaledPod)
}

func (r *PodArgumentor) InjectDecoder(decoder *admission.Decoder) error {
	r.decoder = decoder
	return nil
}

func (r *PodArgumentor) SetupWebhookWithManager(mgr ctrl.Manager) {
	argumentorlog.Info("registering argumentor webhook")
	mgr.GetWebhookServer().
		Register("/mutate-v1-pod", &webhook.Admission{
			Handler: r,
		})
}

func (r *PodArgumentor) Register(h annotation.Handler) {
	r.handlers = append(r.handlers, h)
}
