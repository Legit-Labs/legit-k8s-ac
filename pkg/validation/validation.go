package validation

import (
	"context"
	"fmt"

	"github.com/legit-labs/legit-provenance-verifier/pkg/legit_provenance_verifier"
	"github.com/legit-labs/legit-registry-tools/pkg/legit_registry_tools"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

const (
	LEGIT_PROVENANCE_PREFIX = "legit-provenance"
	PROVENANCE_SIGNING_KEY  = "/attestation-key.pub"
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

func newImage(container corev1.Container, forceDigest bool) (*legit_registry_tools.ImageRef, error) {
	if !forceDigest && !legit_registry_tools.HasDigest(container.Image) {
		return nil, fmt.Errorf("referencing a tagged image (without digest) is not allowed")
	}

	ref, err := legit_registry_tools.NewImageRef(container.Image)
	if err != nil {
		return nil, err
	}

	return ref, nil
}

func invalidPod(reason string) validation {
	return validation{Valid: false, Reason: reason}
}

func (v *Validator) validateProvenance(imageRef *legit_registry_tools.ImageRef) error {
	checks := legit_provenance_verifier.ProvenanceChecks{
		Branch: "main", // TODO from config
	}

	err := legit_provenance_verifier.VerifyRemote(context.Background(), imageRef, PROVENANCE_SIGNING_KEY, checks)
	if err != nil {
		return fmt.Errorf("failed to verify provenance for image [%v]: %v", *imageRef, err)
	}

	return nil
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
	forceDigest := true // TODO from cli
	imagesRefs := make([]legit_registry_tools.ImageRef, 0, len(containers))
	for _, container := range containers {
		imageRef, err := newImage(container, forceDigest)
		if err != nil {
			return invalidPod(fmt.Sprintf("image %v is invalid: %v", container.Name, err)), err
		}
		imagesRefs = append(imagesRefs, *imageRef)
	}

	log := logrus.WithField("pod_name", podName)
	log.Print("images: %v", imagesRefs)

	for _, i := range imagesRefs {
		if err := v.validateProvenance(&i); err != nil {
			return invalidPod(fmt.Sprintf("provenance validation for %v failed: %v", i, err)), err
		}

		log.Printf("image %v was verified for a valid provenance & legit score!", i.Name)
	}

	return validation{Valid: true, Reason: "valid pod"}, nil
}
