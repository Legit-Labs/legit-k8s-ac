package validation

import (
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

// Validator is a container for validation
type Validator struct {
	Logger *logrus.Entry
}

// NewValidator returns an initialised instance of Validator
func NewValidator(logger *logrus.Entry) *Validator {
	return &Validator{Logger: logger}
}

// podValidators is an interface used to group functions validating pods
type podValidator interface {
	Validate(*corev1.Pod) (validation, error)
	Name() string
}

type validation struct {
	Valid  bool
	Reason string
}

// ValidatePod returns true if a pod is valid
func (v *Validator) ValidatePod(pod *corev1.Pod) (validation, error) {
	var podName string
	if pod.Name != "" {
		podName = pod.Name
	} else {
		if pod.ObjectMeta.GenerateName != "" {
			podName = pod.ObjectMeta.GenerateName
		}
	}

	containers := pod.Spec.Containers
	images := make([]string, 0, len(containers))
	for _, container := range containers {
		images = append(images, container.Image)
	}

	log := logrus.WithField("pod_name", podName)
	log.Print("images: %v", images)

	// list of all validations to be applied to the pod
	validations := []podValidator{
		nameValidator{v.Logger},
	}

	// apply all validations
	for _, v := range validations {
		var err error
		vp, err := v.Validate(pod)
		if err != nil {
			return validation{Valid: false, Reason: err.Error()}, err
		}
		if !vp.Valid {
			return validation{Valid: false, Reason: vp.Reason}, err
		}
	}

	return validation{Valid: true, Reason: "valid pod"}, nil
}
