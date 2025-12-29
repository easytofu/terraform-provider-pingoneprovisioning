#!/usr/bin/env bash
# scripts/publish-oci.sh
set -euo pipefail

: "${PROVIDER_NAME:=pingone-propagation}"
: "${NAMESPACE:=easygo}"
: "${VERSION:?set VERSION (e.g. 0.1.0)}"

: "${JFROG_HOST:=easygogroup.jfrog.io}"
: "${OCI_REPO:=tofu-providers-oci}"
: "${JFROG_USERNAME:?set JFROG_USERNAME}"
: "${JFROG_TOKEN:?set JFROG_TOKEN}"

if ! command -v go >/dev/null 2>&1; then
  echo "go not found"
  exit 1
fi

if ! command -v zip >/dev/null 2>&1; then
  echo "zip not found"
  exit 1
fi

if ! command -v oras >/dev/null 2>&1; then
  echo "oras not found"
  exit 1
fi

OCI_REF="${JFROG_HOST}/${OCI_REPO}/${NAMESPACE}/${PROVIDER_NAME}:${VERSION}"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${ROOT_DIR}/dist"
LAYOUT_DIR="${ROOT_DIR}/oci-layout"

rm -rf "${DIST_DIR}" "${LAYOUT_DIR}"
mkdir -p "${DIST_DIR}" "${LAYOUT_DIR}"

targets=(
  "darwin/arm64"
  "darwin/amd64"
  "linux/amd64"
  "linux/arm64"
  "windows/amd64"
  "windows/arm64"
)

layout_tags=()

build_zip() {
  local os="$1"
  local arch="$2"
  local exe_suffix=""

  if [[ "${os}" == "windows" ]]; then
    exe_suffix=".exe"
  fi

  local bin="terraform-provider-${PROVIDER_NAME}_v${VERSION}_${os}_${arch}${exe_suffix}"
  local zipf="terraform-provider-${PROVIDER_NAME}_${VERSION}_${os}_${arch}.zip"

  echo "==> build ${os}/${arch}" >&2
  (cd "${ROOT_DIR}" && GOOS="${os}" GOARCH="${arch}" go build -o "${DIST_DIR}/${bin}") || return 1

  echo "==> zip ${zipf}" >&2
  (cd "${DIST_DIR}" && rm -f "${zipf}" && zip -q "${zipf}" "${bin}") || return 1

  echo "${zipf}"
}

oras_add_target() {
  local os="$1"
  local arch="$2"
  local zip_name="$3"

  local tag="${os}_${arch}"
  local platform="${os}/${arch}"

  echo "==> oras push target ${platform} (layout tag: ${tag})" >&2
  (cd "${DIST_DIR}" && oras push \
    --disable-path-validation \
    --artifact-type application/vnd.opentofu.provider-target \
    --artifact-platform "${platform}" \
    --oci-layout "${LAYOUT_DIR}:${tag}" \
    "${zip_name}:archive/zip")

  layout_tags+=("${tag}")
}

for t in "${targets[@]}"; do
  os="${t%/*}"
  arch="${t#*/}"

  zip_name="$(build_zip "${os}" "${arch}")" || exit $?
  oras_add_target "${os}" "${arch}" "${zip_name}"
done

echo "==> create multi-platform index (tag: ${VERSION})" >&2
oras manifest index create \
  --artifact-type application/vnd.opentofu.provider \
  --oci-layout "${LAYOUT_DIR}:${VERSION}" \
  "${layout_tags[@]}"

echo "==> login ${JFROG_HOST}" >&2
oras login "${JFROG_HOST}" -u "${JFROG_USERNAME}" -p "${JFROG_TOKEN}"

echo "==> push to ${OCI_REF}" >&2
oras cp --from-oci-layout "${LAYOUT_DIR}:${VERSION}" "${OCI_REF}"

echo "==> done: ${OCI_REF}" >&2
