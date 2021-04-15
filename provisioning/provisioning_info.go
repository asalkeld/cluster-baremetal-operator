package provisioning

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"

	configv1 "github.com/openshift/api/config/v1"
	metal3iov1alpha1 "github.com/openshift/cluster-baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/openshift/library-go/pkg/operator/events"
)

type ProvisioningInfo struct {
	Client        kubernetes.Interface
	EventRecorder events.Recorder
	ProvConfig    *metal3iov1alpha1.Provisioning
	Namespace     string
	Images        *Images
	Proxy         *configv1.Proxy
	Scheme        *runtime.Scheme
}
