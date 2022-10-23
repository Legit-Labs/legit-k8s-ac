package legit_registry_tools

import (
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"
)

const (
	shaPrefix       = "sha256"
	shaSeparator    = ":"
	digestSeparator = "@"
	tagSeparator    = ":"
)

type ImageRef struct {
	Name   string
	Tag    string
	Digest string
}

func NewImageRef(ref string) (*ImageRef, error) {
	ref, digest := getImageWithDigest(ref)
	hasDigest := digest != ""
	name, tag := getImageWithTag(ref, hasDigest)

	if !hasDigest {
		var err error
		digest, err = crane.Digest(ref)
		if err != nil {
			return nil, err
		}
	}

	return &ImageRef{
		Name:   name,
		Tag:    tag,
		Digest: digest,
	}, nil
}

func (i *ImageRef) Ref() string {
	ref := i.Name
	if i.Tag != "" {
		ref = fmt.Sprintf("%v:%v", ref, i.Tag)
	}
	if i.Digest != "" {
		ref = fmt.Sprintf("%v@%v", ref, i.Digest)
	}

	return ref
}

func (i *ImageRef) Tagged() bool {
	return i.Tag != ""
}

func DigestToShaValue(digest string) string {
	prefix, pure, found := strings.Cut(digest, shaSeparator)
	if found {
		return pure
	} else {
		// already stripped of the prefix, so the prefix is actually the value
		return prefix
	}
}

func DigestFromShaValue(shaValue string) string {
	prefix := shaPrefix + shaSeparator
	if strings.HasPrefix(shaValue, prefix) {
		return shaValue
	} else {
		return fmt.Sprintf("%v%v", prefix, shaValue)
	}
}

func HasDigest(ref string) bool {
	return strings.Contains(ref, digestSeparator)
}
func HasTag(ref string) bool {
	return strings.Contains(ref, tagSeparator)
}
func SplitByDigest(ref string) (string, string, bool) {
	return strings.Cut(ref, digestSeparator)
}
func SplitByTag(ref string) (string, string, bool) {
	return strings.Cut(ref, tagSeparator)
}

func getImageWithDigest(ref string) (subRef, digest string) {
	subRef, digest, _ = SplitByDigest(ref)
	return
}

func getImageWithTag(ref string, hasDigest bool) (subRef, tag string) {
	var hasTag bool
	subRef, tag, hasTag = SplitByTag(ref)
	if !hasTag && !hasDigest {
		tag = "latest" // default when there is no tag & digest
	}

	return
}
