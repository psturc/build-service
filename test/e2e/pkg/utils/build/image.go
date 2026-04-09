package build

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	ginkgo "github.com/onsi/ginkgo/v2"
	pipeline "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
)

// PipelineImageInfo matches the openshift/oc image info layout used by tests (Config.Config.Labels).
type PipelineImageInfo struct {
	Config *PipelineImageDockerConfig `json:"config"`
}

// PipelineImageDockerConfig is the image config blob (v1 compatibility subset).
type PipelineImageDockerConfig struct {
	Config *PipelineImageContainerConfig `json:"config,omitempty"`
}

// PipelineImageContainerConfig holds container runtime config including labels.
type PipelineImageContainerConfig struct {
	Labels map[string]string `json:"labels,omitempty"`
}

// GetBinaryImage returns the output-image PipelineRun param value, or empty string if unset.
func GetBinaryImage(pr *pipeline.PipelineRun) string {
	for _, p := range pr.Spec.Params {
		if p.Name == "output-image" {
			return p.Value.StringVal
		}
	}
	return ""
}

// ExtractImage extracts a container image filesystem to a new temp directory and returns that path.
func ExtractImage(image string) (string, error) {
	ref, err := name.ParseReference(image)
	if err != nil {
		return "", fmt.Errorf("cannot parse image reference %q: %w", image, err)
	}
	img, err := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return "", fmt.Errorf("pull image %q: %w", image, err)
	}
	tmpDir, err := os.MkdirTemp(os.TempDir(), "eimage-")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}
	ginkgo.GinkgoWriter.Printf("extracting contents of container image %s to dir: %s\n", image, tmpDir)

	rc := mutate.Extract(img)
	defer rc.Close()

	tr := tar.NewReader(rc)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			_ = os.RemoveAll(tmpDir)
			return "", fmt.Errorf("read image tar: %w", err)
		}
		if hdr == nil || hdr.Name == "" {
			continue
		}
		cleanName := filepath.Clean(hdr.Name)
		if cleanName == "." || strings.HasPrefix(cleanName, ".."+string(os.PathSeparator)) {
			continue
		}
		target := filepath.Join(tmpDir, cleanName)
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, hdr.FileInfo().Mode()); err != nil {
				_ = os.RemoveAll(tmpDir)
				return "", fmt.Errorf("mkdir %s: %w", target, err)
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				_ = os.RemoveAll(tmpDir)
				return "", fmt.Errorf("mkdir parent: %w", err)
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, hdr.FileInfo().Mode())
			if err != nil {
				_ = os.RemoveAll(tmpDir)
				return "", fmt.Errorf("create %s: %w", target, err)
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				_ = os.RemoveAll(tmpDir)
				return "", fmt.Errorf("write %s: %w", target, err)
			}
			if err := f.Close(); err != nil {
				_ = os.RemoveAll(tmpDir)
				return "", fmt.Errorf("close %s: %w", target, err)
			}
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				_ = os.RemoveAll(tmpDir)
				return "", fmt.Errorf("mkdir parent: %w", err)
			}
			if err := os.Symlink(hdr.Linkname, target); err != nil {
				_ = os.RemoveAll(tmpDir)
				return "", fmt.Errorf("symlink %s: %w", target, err)
			}
		default:
			// Skip other types (devices, etc.)
		}
	}
	return tmpDir, nil
}

// ImageFromPipelineRun resolves the built image from the output-image param and returns config labels.
func ImageFromPipelineRun(pipelineRun *pipeline.PipelineRun) (*PipelineImageInfo, error) {
	var outputImage string
	for _, parameter := range pipelineRun.Spec.Params {
		if parameter.Name == "output-image" {
			outputImage = parameter.Value.StringVal
		}
	}
	if outputImage == "" {
		return nil, fmt.Errorf("output-image in PipelineRun not found")
	}

	ref, err := name.ParseReference(outputImage)
	if err != nil {
		return nil, fmt.Errorf("parse output image %q: %w", outputImage, err)
	}

	img, err := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return nil, fmt.Errorf("get remote image: %w", err)
	}

	cf, err := img.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("image config: %w", err)
	}

	labels := cf.Config.Labels
	if labels == nil {
		labels = map[string]string{}
	}

	return &PipelineImageInfo{
		Config: &PipelineImageDockerConfig{
			Config: &PipelineImageContainerConfig{
				Labels: labels,
			},
		},
	}, nil
}
