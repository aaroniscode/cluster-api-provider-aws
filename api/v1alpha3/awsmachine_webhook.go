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

package v1alpha3

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
    clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
//	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

// log is for logging in this package.
var foo = logf.Log.WithName("awsmachine-resource")

// TODO, change all doc comments for funtions

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *AWSMachine) ValidateCreate(ctx context.Context, c client.Client) error {
	// Fetch the Machine.
	//machine, err := util.GetOwnerMachine(ctx, c, r.ObjectMeta)
    m := &clusterv1.Machine{}
    key := client.ObjectKey{Name: name, Namespace: namespace}
    if err := c.Get(ctx, key, m); err != nil {
        return nil, err
    }


	if err != nil {
		// todo change error below??
		return apierrors.NewInvalid(GroupVersion.WithKind("AWSMachine").GroupKind(), r.Name, field.ErrorList{
			field.InternalError(nil, errors.Wrap(err, "failed to convert new AWSMachine to unstructured object")),
		})
	}

	foo.Info("Machine is %v", machine)

	if machine.Spec.Version == nil && r.Spec.AMI.ID == nil {
		return apierrors.NewInvalid(GroupVersion.WithKind("AWSMachine").GroupKind(), r.Name, field.ErrorList{
			field.Required(field.NewPath("spec", "ami", "id"),
				"AWSMachine's spec.ami.id is required if Machines's spec.version is not set",
			),
		})
	}
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *AWSMachine) ValidateUpdate(old runtime.Object) error {
	newAWSMachine, err := runtime.DefaultUnstructuredConverter.ToUnstructured(r)
	if err != nil {
		return apierrors.NewInvalid(GroupVersion.WithKind("AWSMachine").GroupKind(), r.Name, field.ErrorList{
			field.InternalError(nil, errors.Wrap(err, "failed to convert new AWSMachine to unstructured object")),
		})
	}
	oldAWSMachine, err := runtime.DefaultUnstructuredConverter.ToUnstructured(old)
	if err != nil {
		return apierrors.NewInvalid(GroupVersion.WithKind("AWSMachine").GroupKind(), r.Name, field.ErrorList{
			field.InternalError(nil, errors.Wrap(err, "failed to convert old AWSMachine to unstructured object")),
		})
	}

	var allErrs field.ErrorList

	newAWSMachineSpec := newAWSMachine["spec"].(map[string]interface{})
	oldAWSMachineSpec := oldAWSMachine["spec"].(map[string]interface{})

	// allow changes to providerID
	delete(oldAWSMachineSpec, "providerID")
	delete(newAWSMachineSpec, "providerID")

	// allow changes to additionalTags
	delete(oldAWSMachineSpec, "additionalTags")
	delete(newAWSMachineSpec, "additionalTags")

	// allow changes to additionalSecurityGroups
	delete(oldAWSMachineSpec, "additionalSecurityGroups")
	delete(newAWSMachineSpec, "additionalSecurityGroups")

	if !reflect.DeepEqual(oldAWSMachineSpec, newAWSMachineSpec) {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec"), "cannot be modified"))
		return apierrors.NewInvalid(
			GroupVersion.WithKind("AWSMachine").GroupKind(),
			r.Name, allErrs)
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *AWSMachine) ValidateDelete() error {
	return nil
}
