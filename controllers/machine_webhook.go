/*
Copyright 2019 The Kubernetes Authors.

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

package controllers

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	infrav1 "sigs.k8s.io/cluster-api-provider-aws/api/v1alpha3"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/cluster-api/controllers/external"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-infrastructure-cluster-x-k8s-io-v1alpha3-machine,mutating=false,failurePolicy=fail,groups=cluster.x-k8s.io,resources=machines,verbs=create;update,versions=v1alpha3,name=providervalidation.machine.infrastructure.cluster.x-k8s.io

// MachineWebhook handles Machine admission webhook requests.
type MachineWebhook struct {
	client  client.Client
	decoder *admission.Decoder
}

// Handle validates Machine admission requests.
func (v *MachineWebhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	machine := &clusterv1.Machine{}

	if err := v.decoder.Decode(req, machine); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if machine.Spec.InfrastructureRef.Kind != "AWSMachine" {
		return admission.Allowed("")
	}

	// Fetch the AWSMachine instance
	object, err := external.Get(ctx, v.client, &machine.Spec.InfrastructureRef, machine.Namespace)
	if err != nil {
		// Allow Machine if AWSMachine isn't available
		if apierrors.IsNotFound(errors.Cause(err)) {
			return admission.Allowed("")
		}
		return admission.Denied(err.Error())
	}

	awsMachine := &infrav1.AWSMachine{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(object.Object, awsMachine); err != nil {
		err = apierrors.NewInvalid(infrav1.GroupVersion.WithKind("AWSMachine").GroupKind(), machine.Name, field.ErrorList{
			field.InternalError(nil, errors.Wrap(err, "failed to convert Machine's InfrastructureRef into AWSMachine")),
		})
		return admission.Denied(err.Error())
	}

	return v.Validate(machine, awsMachine)
}

// Validate todo
func (v *MachineWebhook) Validate(machine *clusterv1.Machine, awsMachine *infrav1.AWSMachine) admission.Response {
	// todo
	if machine.Spec.Version == nil && awsMachine.Spec.AMI.ID == nil {
		err := apierrors.NewInvalid(clusterv1.GroupVersion.WithKind("Machine").GroupKind(), machine.Name, field.ErrorList{
			field.Required(field.NewPath("spec", "version"),
				"Machine's spec.version is required if AWSMachine's spec.ami.id is not set",
			),
		})
		return admission.Denied(err.Error())
	}
	return admission.Allowed("")
}

// InjectClient implements inject.Client.
func (v *MachineWebhook) InjectClient(c client.Client) error {
	v.client = c
	return nil
}

// InjectDecoder implements admission.DecoderInjector.
func (v *MachineWebhook) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}
