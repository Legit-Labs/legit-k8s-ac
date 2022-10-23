package legit_provenance_verifier

import (
	"context"
	"fmt"
	"os"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
	"github.com/legit-labs/legit-registry-tools/pkg/legit_registry_tools"
	"github.com/legit-labs/legit-verify-attestation/pkg/legit_verify_attestation"
)

const (
	LEGIT_PROVENANCE_PREFIX = "legit-provenance"
)

var verifyPayload = legit_verify_attestation.VerifiedTypedPayload[intoto.ProvenanceStatement]

func Verify(ctx context.Context, attestation []byte, keyPath string, digest string, checks ProvenanceChecks) error {
	digest = legit_registry_tools.DigestToShaValue(digest)

	statement, err := verifyPayload(ctx, keyPath, attestation)
	if err != nil {
		return fmt.Errorf("provenance payload verification failed: %v", err)
	}

	if err := legit_verify_attestation.VerifyDigests(statement.Subject, digest); err != nil {
		return fmt.Errorf("provenance digests verification failed: %v", err)
	}

	checker := newProvenanceChecker(statement, checks)
	if err = checker.Check(); err != nil {
		return fmt.Errorf("provenance checks failed: %v", err)
	}

	return nil
}

func VerifyRemote(ctx context.Context, imageRef *legit_registry_tools.ImageRef, keyPath string, checks ProvenanceChecks) error {
	attestation, err := fetchProvenance(imageRef)
	if err != nil {
		return err
	}

	err = Verify(ctx, attestation, keyPath, imageRef.Digest, checks)
	if err != nil {
		return err
	}

	return nil
}

func fetchProvenance(imageRef *legit_registry_tools.ImageRef) ([]byte, error) {
	var attestation []byte

	err := withTmpDir(func(tmpDir string) error {
		attestationPath, err := legit_registry_tools.DownloadAttestation(imageRef.Name, LEGIT_PROVENANCE_PREFIX, tmpDir, imageRef.Digest)
		if err != nil {
			return fmt.Errorf("failed to download attestation: %v", err)
		}

		attestation, err = os.ReadFile(attestationPath)
		if err != nil {
			return fmt.Errorf("failed to read attestation: %v", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return attestation, nil
}

func withTmpDir(foo func(tmpDir string) error) error {
	dir, err := os.MkdirTemp("", "provenance-verification-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	err = foo(dir)
	if err != nil {
		return err
	}

	return nil
}
