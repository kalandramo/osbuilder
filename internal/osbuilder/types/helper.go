package types

import (
	"fmt"
	"path/filepath"

	stringsutil "github.com/onexstack/onexstack/pkg/util/strings"
	"github.com/onexstack/osbuilder/internal/osbuilder/known"
)

func AddGenericPackages(pairs map[string]string, internalPkg string) {
	add := func(dst, tpl string) {
		pairs[dst] = tpl
	}

	add(filepath.Join(internalPkg, "contextx/contextx.go"), "/project/internal/pkg/contextx/contextx.go")
	add(filepath.Join(internalPkg, "contextx/doc.go"), "/project/internal/pkg/contextx/doc.go")
	add(filepath.Join(internalPkg, "known/doc.go"), "/project/internal/pkg/known/doc.go")
	add(filepath.Join(internalPkg, "known/known.go"), "/project/internal/pkg/known/known.go")

	add(filepath.Join(internalPkg, "rid/doc.go"), "/project/internal/pkg/rid/doc.go")
	add(filepath.Join(internalPkg, "rid/example_test.go"), "/project/internal/pkg/rid/example_test.go")
	add(filepath.Join(internalPkg, "rid/rid.go"), "/project/internal/pkg/rid/rid.go")
	add(filepath.Join(internalPkg, "rid/rid_test.go"), "/project/internal/pkg/rid/rid_test.go")
	add(filepath.Join(internalPkg, "rid/salt.go"), "/project/internal/pkg/rid/salt.go")

	add(filepath.Join(internalPkg, "errno/doc.go"), "/project/internal/pkg/errno/doc.go")
	add(filepath.Join(internalPkg, "errno/code.go"), "/project/internal/pkg/errno/code.go")
}

func GenerateKubernetesManifests(pairs map[string]string, meta *Metadata, binaryName string) {
	add := func(dst, tpl string) {
		pairs[dst] = tpl
	}

	if stringsutil.StringIn(meta.DeploymentMethod, []string{known.DeploymentModeDocker, known.DeploymentModeKubernetes}) {
		dockerDir := "/project/build/docker/mb-apiserver"

		switch meta.Image.DockerfileMode {
		case known.DockerfileModeNone:
			add(filepath.Join("build", "docker", binaryName, ".keep"), "/keep.tpl")
		case known.DockerfileModeRuntimeOnly:
			add(filepath.Join("build", "docker", binaryName, "Dockerfile"), fmt.Sprintf("%s/Dockerfile.runtime-only", dockerDir))
		case known.DockerfileModeMultiStage:
			add(filepath.Join("build", "docker", binaryName, "Dockerfile"), fmt.Sprintf("%s/Dockerfile.multi-stage", dockerDir))
		case known.DockerfileModeCombined:
			add(filepath.Join("build", "docker", binaryName, "Dockerfile"), fmt.Sprintf("%s/Dockerfile.multi-stage", dockerDir))
			add(filepath.Join("build", "docker", binaryName, "Dockerfile.runtime-only"), fmt.Sprintf("%s/Dockerfile.runtime-only", dockerDir))
		default:
		}
	}

	if meta.DeploymentMethod == known.DeploymentModeKubernetes {
		manifestDir := "/project/manifests/mb-apiserver"
		add(filepath.Join("manifests", binaryName, "deployment.yaml"), fmt.Sprintf("%s/deployment.yaml", manifestDir))
		add(filepath.Join("manifests", binaryName, "service.yaml"), fmt.Sprintf("%s/service.yaml", manifestDir))
		add(filepath.Join("manifests", binaryName, "create-configmap.sh"), fmt.Sprintf("%s/create-configmap.sh", manifestDir))
		add(filepath.Join("manifests", "nettool.deployment.yaml"), "/project/manifests/nettool.deployment.yaml")
		add(filepath.Join("manifests", "nginx.deployment.yaml"), "/project/manifests/nginx.deployment.yaml")
	}
}
