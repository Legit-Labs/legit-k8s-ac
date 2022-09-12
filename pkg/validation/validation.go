package validation

import (
	"fmt"
	"log"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"
	legit_provenance "github.com/legit-labs/legit-provenance-verifier/legit-provenance"
	legit_score_verifier "github.com/legit-labs/legit-score-verifier/legit-score-verifier"
	registry_tools "github.com/legit-labs/registry-tools/registry-tools"
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

func stripShaPrefix(digest string) string {
	parts := strings.Split(digest, ":")
	if len(parts) != 2 {
		log.Panicf("unexpected digest: %v", digest)
	}
	return parts[1]
}

func (v *Validator) validateProvenance(image ImageData) error {
	imageName := image.Name
	prefix := "provenance"
	dstDir := "/tmp"
	attestationPath, err := registry_tools.DownloadAttestation(imageName, prefix, dstDir, image.Digest)
	if err != nil {
		return fmt.Errorf("failed to download attestation: %v", err)
	}

	keyPath := "/key.pub"
	digest := stripShaPrefix(image.Digest)
	checks := legit_provenance.ProvenanceChecks{
		RepoUrl:   "https://github.com/Legit-Labs/HelloWorld",
		Branch:    "main",
		BuilderId: "https://github.com/legit-labs/legit-provenance-generator/.github/workflows/legit_provenance_generator.yml@refs/tags/v0.1.0",
		Tag:       image.Tag,
	}

	err = legit_provenance.Verify(attestationPath, keyPath, digest, checks)
	if err != nil {
		return err
	}

	return nil
}

func (v *Validator) validateScore(image ImageData) error {
	imageName := image.Name
	prefix := "legit-score"
	dstDir := "/tmp"
	attestationPath, err := registry_tools.DownloadAttestation(imageName, prefix, dstDir, image.Digest)
	if err != nil {
		return fmt.Errorf("failed to download attestation: %v", err)
	}

	keyPath := "/key.pub"
	digest := stripShaPrefix(image.Digest)
	min_score := 6.5
	repo := "TODO repo"
	err = legit_score_verifier.Verify(attestationPath, keyPath, digest, min_score, repo)
	if err != nil {
		return err
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

	for _, i := range images {
		if err := v.validateProvenance(i); err != nil {
			return invalidPod(fmt.Sprintf("provenance validation for %v failed: %v", i, err)), err
		}

		if err := v.validateScore(i); err != nil {
			return invalidPod(fmt.Sprintf("Legit score validation for %v failed: %v", i, err)), err
		}
	}

	return validation{Valid: true, Reason: "valid pod"}, nil
}
