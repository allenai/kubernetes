package specrequirements

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/batch"
	"k8s.io/kubernetes/pkg/apis/extensions"
)

func TestHasContact(t *testing.T) {
	labelsWithContact := make(map[string]string)
	expectedContact := "hodor"
	labelsWithContact["contact"] = expectedContact
	labelsWithContact["app"] = "doorholding"
	podWithContact := api.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "podWithContact", Namespace: "infrastructure", Labels: labelsWithContact},
	}
	resultingContact := HasContact(&podWithContact)
	LabelHelper(expectedContact, resultingContact, t)

	expectedContact = ""
	jobWithoutContact := batch.Job{
		ObjectMeta: metav1.ObjectMeta{Name: "jobWithOutContact", Namespace: "infrastructure"},
	}
	resultingContact = HasContact(&jobWithoutContact)
	LabelHelper(expectedContact, resultingContact, t)
}

func TestPropagateLabel(t *testing.T) {

	// no labels present in template spec
	podTemplate := api.PodTemplateSpec{
		Spec: api.PodSpec{
			Containers: []api.Container{
				{Name: "container1", Image: "image"},
				{Name: "container2", Image: "anotherimage"},
			},
		},
	}
	_, ok := podTemplate.Labels["contact"]
	if ok {
		t.Errorf("Contact label was unexpectedly found in pod spec.")
	}
	expected := "hodor"
	PropagateLabel(&podTemplate, expected)
	actual, err := podTemplate.Labels["contact"]
	if !err {
		t.Errorf("Contact label was not found in pod spec.")
	}
	LabelHelper(expected, actual, t)

	// some labels present in template spec - the spec after calling PropagateLabels should contain the original labels
	// and the new contact label
	labels := make(map[string]string)
	labels["app"] = "train"
	labels["project"] = "infra"
	podTemplateWithSomeLabels := api.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{Labels: labels},
		Spec: api.PodSpec{
			Containers: []api.Container{
				{Name: "container1", Image: "image"},
				{Name: "container2", Image: "anotherimage"},
			},
		},
	}
	_, ok = podTemplateWithSomeLabels.Labels["contact"]
	if ok {
		t.Errorf("Contact label was unexpectedly found in pod spec.")
	}
	expected = "hodor"
	PropagateLabel(&podTemplateWithSomeLabels, expected)
	actual, err = podTemplateWithSomeLabels.Labels["contact"]
	if !err {
		t.Errorf("Contact label was not found in pod spec.")
	}
	LabelHelper(expected, actual, t)
	actual, err = podTemplateWithSomeLabels.Labels["app"]
	if !err {
		t.Errorf("app label was not found in pod spec.")
	}
	LabelHelper("train", actual, t)
	actual, err = podTemplateWithSomeLabels.Labels["project"]
	if !err {
		t.Errorf("project label was not found in pod spec.")
	}
	LabelHelper("infra", actual, t)
}

func LabelHelper(expected string, actual string, t *testing.T) {
	if expected != actual {
		t.Errorf("Label value should have been %s, but was %s.", expected, actual)
	}
}

func TestHasResourceRequests(t *testing.T) {

	// Values for both CPU and memory included
	bothResourcesMap := make(map[api.ResourceName]resource.Quantity)
	bothResourcesMap[api.ResourceCPU] = *resource.NewQuantity(100, resource.DecimalSI)
	bothResourcesMap[api.ResourceMemory] = *resource.NewQuantity(200, resource.DecimalSI)
	bothResourcesRequest := api.ResourceRequirements{Requests: bothResourcesMap}

	// Request for CPU only
	cpuOnlyMap := make(map[api.ResourceName]resource.Quantity)
	cpuOnlyMap[api.ResourceCPU] = *resource.NewQuantity(100, resource.DecimalSI)
	cpuOnlyRequest := api.ResourceRequirements{Requests: cpuOnlyMap}

	// Request for memory only
	memoryOnlyMap := make(map[api.ResourceName]resource.Quantity)
	memoryOnlyMap[api.ResourceMemory] = *resource.NewQuantity(200, resource.DecimalSI)
	memoryOnlyRequest := api.ResourceRequirements{Requests: memoryOnlyMap}

	// All containers have both CPU and memory requests. Should return true.
	twoContainersWithRequests := []api.Container{
		{Name: "first", Image: "image1", Resources: bothResourcesRequest},
		{Name: "second", Image: "image2", Resources: bothResourcesRequest},
	}

	// No containers have any requests. Should return false.
	noContainersHaveRequests := []api.Container{
		{Name: "first", Image: "image1"},
		{Name: "second", Image: "image2"},
	}

	// One container has both CPU and memory requests, the other has no requests. Should return false.
	mixture := []api.Container{
		{Name: "first", Image: "image1", Resources: bothResourcesRequest},
		{Name: "second", Image: "image2"},
	}

	// One container hasn't specified a memory request. Should return false.
	memoryMissing := []api.Container{
		{Name: "first", Image: "image1", Resources: cpuOnlyRequest},
		{Name: "second", Image: "image2"},
	}

	// One container hasn't specified a CPU request. Should return false.
	cpuMissing := []api.Container{
		{Name: "first", Image: "image1"},
		{Name: "second", Image: "image2", Resources: memoryOnlyRequest},
	}

	result := HasResourceRequests(twoContainersWithRequests)
	resourceHelper(true, result, t)
	result = HasResourceRequests(noContainersHaveRequests)
	resourceHelper(false, result, t)
	result = HasResourceRequests(mixture)
	resourceHelper(false, result, t)
	result = HasResourceRequests(memoryMissing)
	resourceHelper(false, result, t)
	result = HasResourceRequests(cpuMissing)
	resourceHelper(false, result, t)
}

func resourceHelper(expected bool, actual bool, t *testing.T) {
	if expected != actual {
		t.Errorf("Unexpected result for resource check. Was %t instead of %t.", actual, expected)
	}
}

func TestGetPodTemplate(t *testing.T) {

	// Test for Jobs
	jobPodTemplateSpec := api.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{Name: "jobpod"},
		Spec: api.PodSpec{
			Containers: []api.Container{
				{Name: "jobcontainer1", Image: "jobimage"},
			},
		},
	}
	job := batch.Job{
		ObjectMeta: metav1.ObjectMeta{Name: "job"},
		Spec: batch.JobSpec{
			Template: jobPodTemplateSpec,
		},
	}
	template, hasTemplate := GetPodTemplate(&job)
	templateGetterHelper(&jobPodTemplateSpec, true, template, hasTemplate, t)

	// Test for Deployments
	deploymentPodTemplateSpec := api.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{Name: "deploymentpod"},
		Spec: api.PodSpec{
			Containers: []api.Container{
				{Name: "deploymentcontainer1", Image: "deploymentimage"},
			},
		},
	}
	deployment := extensions.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "deployment"},
		Spec: extensions.DeploymentSpec{
			Template: deploymentPodTemplateSpec,
		},
	}
	template, hasTemplate = GetPodTemplate(&deployment)
	templateGetterHelper(&deploymentPodTemplateSpec, true, template, hasTemplate, t)

	// Testing resources that do not have PodTemplateSpecs
	// We expect an empty pod template spec for these
	emptyTemplate := &api.PodTemplateSpec{}

	// Test for Pods - Pods contain PodSpecs, there is no PodTemplateSpec.
	pod := api.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod"},
		Spec: api.PodSpec{
			Containers: []api.Container{
				{Name: "deploymentcontainer1", Image: "deploymentimage"},
			},
		},
	}
	template, hasTemplate = GetPodTemplate(&pod)
	templateGetterHelper(emptyTemplate, false, template, hasTemplate, t)

	// Test for Services
	selectors := make(map[string]string)
	selectors["app"] = "search"
	selectors["project"] = "infra"
	service := api.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "service"},
		Spec: api.ServiceSpec {
			Selector: selectors,
		},
	}
	template, hasTemplate = GetPodTemplate(&service)
	templateGetterHelper(emptyTemplate, false, template, hasTemplate, t)


}

func templateGetterHelper(expectedTemplate *api.PodTemplateSpec, expectedPresence bool, actualTemplate *api.PodTemplateSpec, actualPresence bool, t *testing.T){
	if expectedPresence != actualPresence || !reflect.DeepEqual(expectedTemplate, actualTemplate) {
		t.Errorf("Unexpected results when attempting to get pod spec template from resource.\n %+v,\n %t,\n %+v,\n %t", expectedTemplate, expectedPresence, actualTemplate, actualPresence)
	}
}

 //TestAdmission verifies all create requests for pods result in every container's image pull policy
 //set to Always
func TestAdmission(t *testing.T) {

	handler := &specRequirements{}

	// Testing resources created by human.
	humanUser := user.DefaultInfo{Name: "admin"}

	// No contact, contains no pods. Should not be admitted.
	selectors := make(map[string]string)
	selectors["app"] = "search"
	selectors["project"] = "infra"
	noContactService := api.Service {
		ObjectMeta: metav1.ObjectMeta{Name: "service"},
		Spec: api.ServiceSpec {
			Selector: selectors,
		},
	}
	input := admission.NewAttributesRecord(&noContactService, nil, api.Kind("Service").WithVersion("version"), "", "", api.Resource("services").WithVersion("version"), "", admission.Create, &humanUser)
	err := handler.Admit(input)
	if err == nil {
		t.Errorf("A service without a contact label was admitted.")
	}

	// Contains contact but not resources. Should not be admitted.
	labelsWithContact := make(map[string]string)
	labelsWithContact["contact"] = "hodor"
	deploymentWithContactNoResources := extensions.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "deployment", Labels: labelsWithContact},
		Spec: extensions.DeploymentSpec{
			Template: api.PodTemplateSpec{
				Spec: api.PodSpec{
					Containers: []api.Container{
						{Name: "deploymentcontainer1", Image: "deploymentimage"},
					},
				},
			},
		},
	}

	input = admission.NewAttributesRecord(&deploymentWithContactNoResources, nil, api.Kind("Job").WithVersion("version"), "", "", api.Resource("jobs").WithVersion("version"), "", admission.Update, &humanUser)
	err = handler.Admit(input)
	if err == nil {
		t.Errorf("Deployment with contact label but no resources was admitted.")
	}

	// Contains contact and resources. Should be admitted.
	otherLabels := make(map[string]string)
	otherLabels["app"] = "study"
	otherLabels["project"] = "infra"
	bothResourcesMap := make(map[api.ResourceName]resource.Quantity)
	bothResourcesMap[api.ResourceCPU] = *resource.NewQuantity(100, resource.DecimalSI)
	bothResourcesMap[api.ResourceMemory] = *resource.NewQuantity(200, resource.DecimalSI)
	bothResourcesRequest := api.ResourceRequirements{Requests: bothResourcesMap}
	jobWithContactAndResources := batch.Job{
		ObjectMeta: metav1.ObjectMeta{Name: "job", Labels: labelsWithContact},
		Spec: batch.JobSpec{
			Template: api.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: otherLabels},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{Name: "jobcontainer1", Image: "jobimage", Resources: bothResourcesRequest},
					},
				},
			},
		},
	}

	input = admission.NewAttributesRecord(&jobWithContactAndResources, nil, api.Kind("Job").WithVersion("version"), "", "", api.Resource("jobs").WithVersion("version"), "", admission.Create, &humanUser)
    err = handler.Admit(input)
	if err != nil {
		t.Errorf("Job with both contact label and resource requests should have been admitted, but wasn't.")
	}
	// Verify that other pod labels have not been overwritten, and pod has contact label too now
	LabelHelper("study", jobWithContactAndResources.Spec.Template.Labels["app"], t)
	LabelHelper("infra", jobWithContactAndResources.Spec.Template.Labels["project"], t)
	LabelHelper("hodor", jobWithContactAndResources.Spec.Template.Labels["contact"], t)

	// Make sure that nothing goes wrong when the pod template spec doesn't already have at least one label.
	jobWithResourcesAndOnlyContactLabel := batch.Job{
		ObjectMeta: metav1.ObjectMeta{Name: "job2", Labels: labelsWithContact},
		Spec: batch.JobSpec{
			Template: api.PodTemplateSpec{
				Spec: api.PodSpec{
					Containers: []api.Container{
						{Name: "jobcontainer2", Image: "jobimage2", Resources: bothResourcesRequest},
					},
				},
			},
		},
	}

	input = admission.NewAttributesRecord(&jobWithResourcesAndOnlyContactLabel, nil, api.Kind("Job").WithVersion("version"), "", "", api.Resource("jobs").WithVersion("version"), "", admission.Create, &humanUser)
	err = handler.Admit(input)
	if err != nil {
		t.Errorf("Job with both contact label and resource requests should have been admitted, but wasn't.")
	}
	// Verify that the pod has a contact label too now
	LabelHelper("hodor", jobWithResourcesAndOnlyContactLabel.Spec.Template.Labels["contact"], t)

	// ConfigMaps only need to have a contact label, as they do not require pod resources
	cmMap := make(map[string]string)
	cmMap["key1"] = "value1"
	cmMap["key2"] = "value2"
	configMap := api.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "configmap", Labels: labelsWithContact},
		Data: cmMap,
	}
	input = admission.NewAttributesRecord(&configMap, nil, api.Kind("Job").WithVersion("version"), "", "", api.Resource("jobs").WithVersion("version"), "", admission.Update, &humanUser)
	err = handler.Admit(input)
	if err != nil {
		t.Errorf("ConfigMap with contact label should have been admitted, but wasn't.")
	}

	// Testing admission when the humanUser is not human (i.e. Kubernetes)
	// Service with no contact label
	kubeUser := user.DefaultInfo{Name: "kubelet"}
	input = admission.NewAttributesRecord(&noContactService, nil, api.Kind("Service").WithVersion("version"), "", "", api.Resource("services").WithVersion("version"), "", admission.Create, &kubeUser)
	if err != nil {
		t.Errorf("The kubelet user should have been able to create a service without a contact label, but was unable to.")
	}

	// Deployment with contact label but no resource requests
	input = admission.NewAttributesRecord(&deploymentWithContactNoResources, nil, api.Kind("Job").WithVersion("version"), "", "", api.Resource("jobs").WithVersion("version"), "", admission.Update, &kubeUser)
	if err != nil {
		t.Errorf("The kubelet user should have been able to update a deployment with no resource requests, but was unable to.")
	}
}
