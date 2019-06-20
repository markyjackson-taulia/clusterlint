package basic

import (
	"fmt"

	"github.com/digitalocean/clusterlint/checks"
	"github.com/digitalocean/clusterlint/kube"
	"github.com/docker/distribution/reference"
	corev1 "k8s.io/api/core/v1"
)

func init() {
	checks.Register(&fullyQualifiedImageCheck{})
}

type fullyQualifiedImageCheck struct{}

// Name returns a unique name for this check.
func (fq *fullyQualifiedImageCheck) Name() string {
	return "fully-qualified-image"
}

// Groups returns a list of group names this check should be part of.
func (fq *fullyQualifiedImageCheck) Groups() []string {
	return []string{"basic"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (fq *fullyQualifiedImageCheck) Description() string {
	return "Checks if containers have fully qualified image names"
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (fq *fullyQualifiedImageCheck) Run(objects *kube.Objects) ([]error, []error, error) {
	var warnings, errors []error

	for _, pod := range objects.Pods.Items {
		podName := pod.GetName()
		namespace := pod.GetNamespace()
		w, e := checkImage(pod.Spec.Containers, podName, namespace)
		warnings = append(warnings, w...)
		errors = append(errors, e...)
		w, e = checkImage(pod.Spec.InitContainers, podName, namespace)
		warnings = append(warnings, w...)
		errors = append(errors, e...)
	}

	return warnings, errors, nil
}

// checkImage checks if the image name is fully qualified
// Adds a warning if the container does not use a fully qualified image name
func checkImage(containers []corev1.Container, podName string, namespace string) ([]error, []error) {
	var w, e []error
	for _, container := range containers {
		value, err := reference.ParseAnyReference(container.Image)
		if err != nil {
			e = append(e, fmt.Errorf("[Error] Malformed image name for container '%s' in pod '%s' in namespace '%s'", container.Name, podName, namespace))
		} else {
			if value.String() != container.Image {
				w = append(w, fmt.Errorf("[Best Practice] Use fully qualified image for container '%s' in pod '%s' in namespace '%s'", container.Name, podName, namespace))
			}
		}
	}
	return w, e
}