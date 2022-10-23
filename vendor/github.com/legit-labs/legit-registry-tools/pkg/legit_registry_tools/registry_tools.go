package legit_registry_tools

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"
)

func TaggedRefToDigestedRef(ref string) (string, error) {
	digest, err := crane.Digest(ref)
	if err != nil {
		return "", err
	}

	name := strings.Split(ref, ":")[0]
	return name + "@" + digest, nil
}

func UploadFile(name string, path string, ref string) error {
	m := make(map[string][]byte)
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	m[name] = content
	image, err := crane.Image(m)
	if err != nil {
		return err
	}

	err = crane.Push(image, ref)
	if err != nil {
		return err
	}

	return nil
}

func PullSingleLayerIntoDir(ref string, dstDir string) error {
	i, err := crane.Pull(ref)
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "tar-extraction-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	tmpTarName := filepath.Join(tmpDir, "image.tar")
	w, err := os.OpenFile(tmpTarName, os.O_WRONLY|os.O_CREATE, os.FileMode(0600))
	if err != nil {
		return err
	}
	defer w.Close() // XXX: no need to remove, handled by removing the temporary directory

	err = crane.Export(i, w)
	if err != nil {
		return err
	}

	r, err := os.OpenFile(tmpTarName, os.O_RDONLY, os.FileMode(0400))
	if err != nil {
		return err
	}
	defer r.Close()

	err = untar(dstDir, r)
	if err != nil {
		return err
	}

	return nil
}

func UploadAttestation(image string, prefix string, path string, digest string) error {
	info := NewAttestationInfo(image, prefix, digest)
	return UploadFile(info.Name, path, info.Ref)
}

func DownloadAttestation(image string, prefix string, dstDir string, digest string) (path string, err error) {
	info := NewAttestationInfo(image, prefix, digest)
	err = PullSingleLayerIntoDir(info.Ref, dstDir)
	if err != nil {
		return
	}

	path = filepath.Join(dstDir, info.Name)
	return
}
