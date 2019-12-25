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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-infrastructure-cluster-x-k8s-io-v1alpha3-machine,mutating=false,failurePolicy=fail,groups=cluster.x-k8s.io,resources=machines,verbs=create;update,versions=v1alpha3,name=providervalidation.machine.infrastructure.cluster.x-k8s.io

// MachineValidator validates Machines
type MachineValidator struct {
	client  client.Client
	decoder *admission.Decoder
}

// Handle validates Machine admission requests.
func (v *MachineValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	m := &clusterv1.Machine{}

	if err := v.decoder.Decode(req, m); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if m.Spec.InfrastructureRef.Kind == "AWSMachine" && m.Spec.Version == nil {
		err := apierrors.NewInvalid(clusterv1.GroupVersion.WithKind("Machine").GroupKind(), m.Name, field.ErrorList{
			field.Required(field.NewPath("spec", "version"), "is required"),
		})
		return admission.Denied(err.Error())
	}

	return admission.Allowed("")
}

// InjectClient implements inject.Client.
func (v *MachineValidator) InjectClient(c client.Client) error {
	v.client = c
	return nil
}

// InjectDecoder implements admission.DecoderInjector.
func (v *MachineValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}
