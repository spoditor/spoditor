package internal

import (
	"context"
	"net/http"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=fail,sideEffects=None,groups="",resources=pods,verbs=create;update,versions=v1,name=mpod.ssarg.io,admissionReviewVersions={v1,v1beta1}

// log is for logging in this package.
var argumentorlog = logf.Log.WithName("argumentor-resource")

type PodArgumentor struct {
	decoder *admission.Decoder
}

func (r *PodArgumentor) Handle(_ context.Context, request admission.Request) admission.Response {
	pod := &v1.Pod{}
	err := r.decoder.Decode(request, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// mutate the fields in pod

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(request.Object.Raw, marshaledPod)
}

func (r *PodArgumentor) SetupWebhookWithManager(mgr ctrl.Manager) {
	argumentorlog.Info("registering argumentor webhook")
	mgr.GetWebhookServer().
		Register("/mutate-v1-pod", &webhook.Admission{
			Handler: &PodArgumentor{},
		})
}
