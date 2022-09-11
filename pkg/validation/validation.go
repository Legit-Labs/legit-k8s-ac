package validation

import (
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"
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

type ImageData struct {
	Name   string
	Tag    string
	Digest string
}

func getImageWithDigest(ref string) (name, digest string) {
	parts := strings.Split(ref, "@")
	name = parts[0]
	digest = parts[1]
	return
}

func getImageInfo(ref string) (name, tag, digest string, err error) {
	if strings.Contains(ref, "@") { // referenced by digest
		name, digest = getImageWithDigest(ref)
		return
	}
	if !strings.Contains(ref, ":") { // no tag defaults to latest
		ref += ":latest"
	}

	parts := strings.Split(ref, ":")
	name = parts[0]
	tag = parts[1]
	digest, err = crane.Digest(ref)
	if err != nil {
		err = fmt.Errorf("failed to get digest for ref %v: %v", ref, err)
		return
	}

	return
}

func newImage(container corev1.Container, allowTagged bool) (*ImageData, error) {
	name, tag, digest, err := getImageInfo(container.Image)
	if err != nil {
		return nil, err
	}

	if !allowTagged && tag != "" {
		return nil, fmt.Errorf("referencing a tagged image is not allowed (%v)", tag)
	}

	return &ImageData{
		Name:   name,
		Tag:    tag,
		Digest: digest,
	}, nil
}

func invalidPod(reason string) validation {
	return validation{Valid: false, Reason: reason}
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
	allowTagged := true // TODO from cli
	images := make([]ImageData, 0, len(containers))
	for _, container := range containers {
		image, err := newImage(container, allowTagged)
		if err != nil {
			return invalidPod(fmt.Sprintf("image %v is invalid: %v", container.Name, err)), err
		}
		images = append(images, *image)
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
			return invalidPod(err.Error()), err
		}
		if !vp.Valid {
			return invalidPod(vp.Reason), err
		}
	}

	return validation{Valid: true, Reason: "valid pod"}, nil
}
