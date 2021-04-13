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

package provisioning

import (
	"context"
	"os"

	admissionregistration "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"

	osconfigv1 "github.com/openshift/api/config/v1"
	osclientset "github.com/openshift/client-go/config/clientset/versioned"
	"github.com/openshift/library-go/pkg/config/clusteroperator/v1helpers"
	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
)

func EnableValidatingWebhook(info *ProvisioningInfo) error {
	ignore := admissionregistration.Ignore
	noSideEffects := admissionregistration.SideEffectClassNone
	instance := &admissionregistration.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster-baremetal-validating-webhook-configuration",
			Annotations: map[string]string{
				"include.release.openshift.io/self-managed-high-availability": "true",
				"include.release.openshift.io/single-node-developer":          "true",
				"service.beta.openshift.io/inject-cabundle":                   "true",
			},
		},
		Webhooks: []admissionregistration.ValidatingWebhook{
			{
				ClientConfig: admissionregistration.WebhookClientConfig{
					CABundle: []byte("Cg=="),
					Service: &admissionregistration.ServiceReference{
						Name:      "cluster-baremetal-webhook-service",
						Namespace: info.Namespace,
						Path:      pointer.StringPtr("/validate-metal3-io-v1alpha1-provisioning"),
					},
				},
				SideEffects:             &noSideEffects,
				FailurePolicy:           &ignore,
				AdmissionReviewVersions: []string{"v1", "v1beta1"},
				Name:                    "vprovisioning.kb.io",
				Rules: []admissionregistration.RuleWithOperations{
					{
						Operations: []admissionregistration.OperationType{
							admissionregistration.Create,
							admissionregistration.Update,
						},
						Rule: admissionregistration.Rule{
							Resources:   []string{"provisionings"},
							APIGroups:   []string{"metal3.io"},
							APIVersions: []string{"v1alpha1"},
						},
					},
				},
			},
		},
	}
	// we might not have a baremetalCR (when disabled), so we have no where to store
	// the expectedGeneration, so just fake it.
	expectedGeneration := int64(0)
	_, _, err := resourceapply.ApplyValidatingWebhookConfiguration(info.Client.AdmissionregistrationV1(), info.EventRecorder, instance, expectedGeneration)
	if err != nil {
		return err
	}

	return nil //(&metal3iov1alpha1.Provisioning{}).SetupWebhookWithManager(mgr)
}

func WebhookDependenciesReady(client osclientset.Interface) bool {
	co, err := client.ConfigV1().ClusterOperators().Get(context.Background(), "service-ca", metav1.GetOptions{})
	if err != nil {
		return false
	}

	for condName, condVal := range map[osconfigv1.ClusterStatusConditionType]osconfigv1.ConditionStatus{
		osconfigv1.OperatorDegraded:    osconfigv1.ConditionFalse,
		osconfigv1.OperatorProgressing: osconfigv1.ConditionFalse,
		osconfigv1.OperatorAvailable:   osconfigv1.ConditionTrue} {
		if !v1helpers.IsStatusConditionPresentAndEqual(co.Status.Conditions, condName, condVal) {
			klog.Infof("WebhookDependenciesReady: service-ca not ready %s!=%s", condName, condVal)
			return false
		}
	}

	for _, fname := range []string{
		"/etc/cluster-baremetal-operator/tls/tls.key",
		"/etc/cluster-baremetal-operator/tls/tls.crt",
	} {
		_, err := os.Stat(fname)
		if err != nil {
			klog.Infof("WebhookDependenciesReady: file does not exist %s", fname)
			return false
		}
	}
	klog.Info("WebhookDependenciesReady: everything ready for webhooks")
	return true
}
