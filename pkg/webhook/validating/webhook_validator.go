package validating

import (
	"context"
	"fmt"
	"net/http"

	kedav1alpha1 "github.com/kedacore/keda-olm-operator/pkg/apis/keda/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	// Allowed Name and Namespace of KedaController resource
	kedaControllerResourceName      = "keda"
	kedaControllerResourceNamespace = "keda"
)

var log = logf.Log.WithName("webhook_validator")

// USE .../kubernetes-sigs/controller-tools/.run-controller-gen.sh webhook paths=./pkg/webhook/... output:dir=./deploy

// +kubebuilder:webhook:path=/validate-keda-k8s-io-v1alpha-kedacontroller,mutating=false,failurePolicy=ignore,groups="keda.k8s.io",resources=kedacontrollers,verbs=create;update,versions=v1alpha,name=kedacontroller.keda.k8s.io

// KedaControllerValidator validates KedaController
type KedaControllerValidator struct {
	Client  client.Client
	decoder *admission.Decoder
}

// podValidator admits a pod iff a specific annotation exists.
func (v *KedaControllerValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	kc := &kedav1alpha1.KedaController{}

	err := v.decoder.Decode(req, kc)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	allowed, reason, err := v.validate(ctx, kc)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	if allowed {
		return admission.Allowed(reason)
	} else {
		return admission.Denied(reason)
	}
}

// podValidator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (v *KedaControllerValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}

// validate
func (v *KedaControllerValidator) validate(ctx context.Context, kc *kedav1alpha1.KedaController) (allowed bool, reason string, err error) {
	stages := []func(context.Context, *kedav1alpha1.KedaController) (bool, string, error){
		v.validateNameAndNamespace,
		v.validateUniqueness,
	}
	for _, stage := range stages {
		allowed, reason, err = stage(ctx, kc)
		if len(reason) > 0 {
			if err != nil {
				log.Error(err, reason)
			} else {
				log.Info(reason)
			}
		}
		if !allowed {
			return
		}
	}
	return
}

// validate required name and namespace
func (v *KedaControllerValidator) validateNameAndNamespace(ctx context.Context, kc *kedav1alpha1.KedaController) (bool, string, error) {
	if kc.Name != kedaControllerResourceName || kc.Namespace != kedaControllerResourceNamespace {
		return false, fmt.Sprintf("The KedaController resource needs to be created in namespace %s with name %s, otherwise it will be ignored", kedaControllerResourceNamespace, kedaControllerResourceName), nil
	}
	return true, "", nil
}

// validate this is the only KedaController in this namespace
func (v *KedaControllerValidator) validateUniqueness(ctx context.Context, kc *kedav1alpha1.KedaController) (bool, string, error) {
	list := &kedav1alpha1.KedaControllerList{}
	if err := v.Client.List(ctx, list, &client.ListOptions{Namespace: kc.Namespace}); err != nil {
		return false, "Unable to list KedaControllers", err
	}
	for _, v := range list.Items {
		if kc.Name != v.Name {
			return false, "Only one KedaControllers allowed per namespace", nil
		}
	}
	return true, "", nil
}
