package webhook

import (
	"github.com/kedacore/keda-olm-operator/pkg/webhook/validating"

	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var log = logf.Log.WithName("webhook")

// AddToManagerFuncs is a list of functions to add all Webhooks to the Manager
var AddToManagerFuncs []func(manager.Manager) error

// AddToManager adds all Webhooks to the Manager
func AddToManager(m manager.Manager) error {

	log.Info("Starting the Webhook Server")

	// Setup webhooks
	hookServer := m.GetWebhookServer()
	hookServer.CertDir = "/certs"
	hookServer.Port = 6443

	log.Info("Registering webhooks to the Webhook Server")
	hookServer.Register("/validate-keda-k8s-io-v1alpha-kedacontroller", &webhook.Admission{Handler: &validating.KedaControllerValidator{Client: m.GetClient()}})

	return nil
}
