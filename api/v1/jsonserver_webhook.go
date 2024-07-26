/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"encoding/json"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"strings"
)

// log is for logging in this package.
var jsonserverlog = logf.Log.WithName("jsonserver-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *JsonServer) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-example-com-v1-jsonserver,mutating=true,failurePolicy=fail,sideEffects=None,groups=example.com,resources=jsonservers,verbs=create;update,versions=v1,name=mjsonserver.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &JsonServer{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *JsonServer) Default() {
	jsonserverlog.Info("default", "name", r.Name)

	if r.Spec.Replicas == nil {
		defaultReplicas := int32(1)
		r.Spec.Replicas = &defaultReplicas
	}
}

// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-example-com-v1-jsonserver,mutating=false,failurePolicy=fail,sideEffects=None,groups=example.com,resources=jsonservers,verbs=create;update,versions=v1,name=vjsonserver.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &JsonServer{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *JsonServer) ValidateCreate() (admission.Warnings, error) {
	jsonserverlog.Info("validate create", "name", r.Name)

	return nil, r.validateJsonServer()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *JsonServer) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	jsonserverlog.Info("validate update", "name", r.Name)

	return nil, r.validateJsonServer()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *JsonServer) ValidateDelete() (admission.Warnings, error) {
	jsonserverlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

func (r *JsonServer) validateJsonServer() error {
	if !strings.HasPrefix(r.Name, "app-") {
		err := field.Invalid(field.NewPath("metadata").Child("name"), r.Name, "metadata.name must start with 'app-'")
		return err
	}

	if !isValidJSON(r.Spec.JsonConfig) {
		err := field.Invalid(field.NewPath("spec").Child("jsonConfig"), r.Spec.JsonConfig, "spec.jsonConfig is not a valid json object")
		return err
	}

	if *r.Spec.Replicas < 0 {
		err := field.Invalid(field.NewPath("spec").Child("replicas"), r.Spec.Replicas, "spec.replicas must be a non-negative integer")
		return err
	}

	return nil
}

func isValidJSON(jsonStr string) bool {
	var js map[string]interface{}

	return json.Unmarshal([]byte(jsonStr), &js) == nil
}
