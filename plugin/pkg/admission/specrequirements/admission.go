/*
Copyright 2015 The Kubernetes Authors.

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

// The SpecRequirements package contains an admission controller that ensures that resources created or updated by
// human cluster users meet two criteria:
// - they have a contact label
// - if pods will be created, they have resource requirements specified
// Resources that do not meet those criteria will not be created/updated
// In addition, the admission controller propagates an existing contact label to pods, if the main resource being
// handled is not a pod but contains pods (e.g. deployments or jobs)
package specrequirements

import (
	"io"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/kubernetes/pkg/api"
)

func init() {
	admission.RegisterPlugin("SpecRequirements", func(config io.Reader) (admission.Interface, error) {
		return NewSpecRequirements(), nil
	})
}

type specRequirements struct {
	*admission.Handler
}

//Admit returns nil if the resource meets the criteria for creation/updating, and an error otherwise.
func (a *specRequirements) Admit(attributes admission.Attributes) (err error) {
	// We only want to filter resources created by human cluster users.
	if !strings.Contains(attributes.GetUserInfo().GetName(), "admin") {
		return nil
	}

	// Check for a contact label
	contact := HasContact(attributes.GetObject())

	// Stop the resource update/creation if there is not contact label
	if contact == "" {
		mess := fmt.Sprintf("Cannot %s this resource. It does not have a valid contact label. Please add one " +
			"and try again.", strings.ToLower(string(attributes.GetOperation())))
		return admission.NewForbidden(attributes, errors.New(mess))
	}

	// Get the pod template spec, if the resource has one
	// We need this to propagate the contact label, and to check whether resource requests have been specified.
	template, hasTemplate := GetPodTemplate(attributes.GetObject())

	if hasTemplate {
		PropagateLabel(template, contact)
	}

	// Only resources with pod specs in them need to specify resource requirements.
	if attributes.GetKind().Kind != "Pod" && !hasTemplate {
		return nil
	}

	// The resource is either a pod, or a resource that contains a pod spec.
	// Find out whether resource requests are specified.
	requestsPresent := true
	var containers []api.Container
	if attributes.GetKind().Kind == "Pod" {
		podSpec := attributes.GetObject().(*api.Pod).Spec
		containers = append(podSpec.InitContainers, podSpec.Containers...)
		requestsPresent = HasResourceRequests(containers)
	} else if hasTemplate {
		containers = append(template.Spec.InitContainers, template.Spec.Containers...)
		requestsPresent = HasResourceRequests(containers)
	}

	// If there is at least one CPU or memory request missing in a container, stop the resource from getting
	// updated or created.
	if !requestsPresent {
		mess := fmt.Sprintf("Cannot %s this resource. Some containers do not have resource requests " +
			"specified. Please add resource requests to every container in the pod spec and try again.",
			strings.ToLower(string(attributes.GetOperation())))
		return admission.NewForbidden(attributes, errors.New(mess))
	}

	return nil
}

//Returns true if the given object has a contact label - not case-sensitive - in its metadata, false otherwise.
func HasContact(resource runtime.Object) string {
	accessor := meta.NewAccessor()
	labels, _ := accessor.Labels(resource)
	for key, value := range labels {
		if strings.ToLower(key) == "contact" {
			return value
		}
	}
	return ""
}


//This adds a contact label to the given pod template spec. This is used in the case of resources that contain pods
//but are not pods themselves (e.g. deployments).
func PropagateLabel(template *api.PodTemplateSpec, contact string) {
	if template.Labels == nil {
		template.Labels = make(map[string]string)
	}
	template.Labels["contact"] = contact
}

//Returns true if all the given containers have memory and cpu resource requests set, false otherwise.
func HasResourceRequests(containers []api.Container) bool {
	requestsPresent := true

	//containers := append(template.Spec.Containers, template.Spec.InitContainers...)
	for _, container := range containers{
		requests := container.Resources.Requests
		_, mem := requests[api.ResourceMemory]
		_, cpu := requests[api.ResourceCPU]
		requestsPresent = requestsPresent && mem && cpu
	}

	return requestsPresent

}

//This returns the PodTemplateSpec part of the given resource and true if the resource has a PodTemplateSpec part.
//It returns an empty PodTemplateSpec and false otherwise.
func GetPodTemplate(resource runtime.Object) (*api.PodTemplateSpec, bool) {

	emptyTemplate := api.PodTemplateSpec{}

	possibleSpec := reflect.ValueOf(resource).Elem().FieldByName("Spec")
	if !possibleSpec.IsValid() {
		return &emptyTemplate, false
	}

	spec := possibleSpec.Addr().Interface()
	possibleTemplate := reflect.ValueOf(spec).Elem().FieldByName("Template")
	if !possibleTemplate.IsValid() {
		return &emptyTemplate, false
	}

	podTemplate := possibleTemplate.Addr().Interface().(*api.PodTemplateSpec)
	return podTemplate, true

}

func NewSpecRequirements() admission.Interface {
	return &specRequirements{
		Handler: admission.NewHandler(admission.Create, admission.Update),
	}
}

