/*
Copyright 2025.

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

package operator

import (
	"slices"

	operatorv1beta1 "github.com/openstack-k8s-operators/openstack-operator/apis/operator/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// custom struct types for rendering the operator gotemplates

// Operator -
type Operator struct {
	Name       string
	Namespace  string
	Deployment Deployment
}

// Deployment -
type Deployment struct {
	Replicas      *int32
	Manager       Container
	KubeRbacProxy Container
}

// Container -
type Container struct {
	Image     string
	Env       []corev1.EnvVar
	Resources Resource
}

// Resource -
type Resource struct {
	Requests *ResourceList
	Limits   *ResourceList
}

// ResourceList -
type ResourceList struct {
	CPU    string // using string here instead of resource.Quantity since this can not be direct used in the gotemplate
	Memory string
}

func cpuQuantity(milli int64) *resource.Quantity {
	q := resource.NewMilliQuantity(milli, resource.DecimalSI)
	return q
}

func memQuantity(mega int64) *resource.Quantity {
	q := resource.NewQuantity(mega*1024*1024, resource.BinarySI)
	return q
}

func HasOverrides(operatorOverrides []operatorv1beta1.OperatorSpec, operatorName string) *operatorv1beta1.OperatorSpec {
	// validate of operatorName is in the list of operatorOverrides
	f := func(c operatorv1beta1.OperatorSpec) bool {
		return c.Name == operatorName
	}
	idx := slices.IndexFunc(operatorOverrides, f)
	if idx >= 0 {
		return &operatorOverrides[idx]
	}

	return nil
}

func SetOverrides(opOvr operatorv1beta1.OperatorSpec, op *Operator) {
	if opOvr.Replicas != nil {
		op.Deployment.Replicas = opOvr.Replicas
	}
	if opOvr.ControllerManager.Resources.Limits != nil {
		if op.Deployment.Manager.Resources.Limits == nil {
			op.Deployment.Manager.Resources.Limits = &ResourceList{}
		}
		if opOvr.ControllerManager.Resources.Limits.Cpu() != nil && opOvr.ControllerManager.Resources.Limits.Cpu().Value() > 0 {
			op.Deployment.Manager.Resources.Limits.CPU = opOvr.ControllerManager.Resources.Limits.Cpu().String()
		}
		if opOvr.ControllerManager.Resources.Limits.Memory() != nil && opOvr.ControllerManager.Resources.Limits.Memory().Value() > 0 {
			op.Deployment.Manager.Resources.Limits.Memory = opOvr.ControllerManager.Resources.Limits.Memory().String()
		}
	}
	if opOvr.ControllerManager.Resources.Requests != nil {
		if op.Deployment.Manager.Resources.Requests == nil {
			op.Deployment.Manager.Resources.Requests = &ResourceList{}
		}
		if opOvr.ControllerManager.Resources.Requests.Cpu() != nil && opOvr.ControllerManager.Resources.Requests.Cpu().Value() > 0 {
			op.Deployment.Manager.Resources.Requests.CPU = opOvr.ControllerManager.Resources.Requests.Cpu().String()
		}
		if opOvr.ControllerManager.Resources.Requests.Memory() != nil && opOvr.ControllerManager.Resources.Requests.Memory().Value() > 0 {
			op.Deployment.Manager.Resources.Requests.Memory = opOvr.ControllerManager.Resources.Requests.Memory().String()
		}
	}
}

func GetOperator(operators []Operator, name string) (int, Operator) {
	f := func(c Operator) bool {
		return c.Name == name
	}
	idx := slices.IndexFunc(operators, f)
	if idx >= 0 {
		return idx, operators[idx]
	}

	return idx, Operator{}
}
