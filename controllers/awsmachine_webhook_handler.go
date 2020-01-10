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

	"k8s.io/api/admission/v1beta1"
	infrav1 "sigs.k8s.io/cluster-api-provider-aws/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:verbs=create;update,path=/validate-infrastructure-cluster-x-k8s-io-v1alpha3-awsmachine,mutating=false,failurePolicy=fail,groups=infrastructure.cluster.x-k8s.io,resources=awsmachines,versions=v1alpha3,name=validation.awsmachine.infrastructure.cluster.x-k8s.io

// NOTE, due to: https://github.com/kubernetes-sigs/controller-runtime/issues/711

// AWSMachineWebhook handles AWSMachine admission webhook requests.
type AWSMachineWebhook struct {
	client  client.Client
	decoder *admission.Decoder
}

// Handle validates AWSMachine admission requests.
func (v *AWSMachineWebhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	awsmachine := &infrav1.AWSMachine{}

	if req.Operation == v1beta1.Create {
		if err := v.decoder.Decode(req, awsmachine); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		if err := awsmachine.ValidateCreate(ctx, v.client); err != nil {
			return admission.Denied(err.Error())
		}
	}

	if req.Operation == v1beta1.Update {
		oldObj := awsmachine.DeepCopyObject()

		if err := v.decoder.DecodeRaw(req.Object, awsmachine); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		if err := v.decoder.DecodeRaw(req.OldObject, oldObj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		if err := awsmachine.ValidateUpdate(oldObj); err != nil {
			return admission.Denied(err.Error())
		}
	}

	if req.Operation == v1beta1.Delete {
		// In reference to PR: https://github.com/kubernetes/kubernetes/pull/76346
		// OldObject contains the object being deleted
		if err := v.decoder.DecodeRaw(req.OldObject, awsmachine); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		if err := awsmachine.ValidateDelete(); err != nil {
			return admission.Denied(err.Error())
		}
	}

	return admission.Allowed("")
}

// InjectClient implements inject.Client.
func (v *AWSMachineWebhook) InjectClient(c client.Client) error {
	v.client = c
	return nil
}

// InjectDecoder implements admission.DecoderInjector.
func (v *AWSMachineWebhook) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}
