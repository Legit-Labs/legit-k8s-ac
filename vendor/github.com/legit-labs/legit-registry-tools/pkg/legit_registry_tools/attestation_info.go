package legit_registry_tools

import (
	"fmt"
	"strings"
)

const (
	attestationExt = "intoto.jsonl"
)

func DigestToLabel(digest string) string {
	// replaces sha256:XXX to sha256-XXX
	return strings.Replace(digest, ":", "-", 1)
}

func AttestationRef(image string, prefix string, digest string) string {
	filename := AttestationName(DigestToLabel(digest))
	return fmt.Sprintf("%v:%v-%v", image, prefix, filename)
}

func AttestationName(prefix string) string {
	return fmt.Sprintf("%v.%v", prefix, attestationExt)
}

type AttestationInfo struct {
	Name string
	Ref  string
}

func NewAttestationInfo(image string, prefix string, digest string) AttestationInfo {
	return AttestationInfo{
		Name: AttestationName(prefix),
		Ref:  AttestationRef(image, prefix, digest),
	}
}
