/*

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

package v1alpha1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func (r *Provisioning) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// see https://book.kubebuilder.io/cronjob-tutorial/webhook-implementation.html
// +kubebuilder:webhook:verbs=create;update,path=/validate-metal3-io-v1alpha1-provisioning,mutating=false,failurePolicy=ignore,groups=metal3.io,resources=provisionings,versions=v1alpha1,name=vprovisioning.kb.io

// https://golangbyexample.com/go-check-if-type-implements-interface/
var _ webhook.Validator = &Provisioning{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Provisioning) ValidateCreate() error {
	klog.Info("validate create", "name", r.Name)

	if r.Name != ProvisioningSingletonName {
		return fmt.Errorf("Provisioning object is a singleton and must be named \"%s\"", ProvisioningSingletonName)
	}

	return r.ValidateBaremetalProvisioningConfig()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Provisioning) ValidateUpdate(old runtime.Object) error {
	klog.Info("validate update", "name", r.Name)
	return r.ValidateBaremetalProvisioningConfig()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Provisioning) ValidateDelete() error {
	klog.Info("validate delete", "name", r.Name)
	return nil
}
