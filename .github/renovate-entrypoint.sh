#!/bin/bash
set -euo pipefail

# Single source of truth for the dagger version is the repo's dagger.json.
# The runner's repo checkout is mounted at /workspace via docker-volumes
# in renovate.yaml.
DAGGER_VERSION="$(jq -r '.engineVersion' /workspace/dagger.json)"
if [ -z "${DAGGER_VERSION}" ] || [ "${DAGGER_VERSION}" = "null" ]; then
  echo "engineVersion is missing from /workspace/dagger.json" >&2
  exit 1
fi

echo "Installing dagger ${DAGGER_VERSION}"
curl -fsSL https://dl.dagger.io/dagger/install.sh \
  | DAGGER_VERSION="${DAGGER_VERSION}" BIN_DIR=/usr/local/bin sh

# Grant the renovate-running user access to the host Docker socket so
# `dagger develop` can spawn the engine container. The socket is mounted
# by renovatebot/github-action when mount-docker-socket: true is set.
if [ -S /var/run/docker.sock ]; then
  DOCKER_GID="$(stat -c '%g' /var/run/docker.sock)"
  groupadd -fg "${DOCKER_GID}" docker || true
  usermod -aG docker ubuntu
fi

groupadd -r nixbld

mkdir -m 0755 /nix
chown -R ubuntu:ubuntu /nix

for n in $(seq 1 10); do
  useradd -c "Nix build user $n" -d /var/empty -g nixbld -G nixbld -M -N -r -s "$(command -v nologin)" "nixbld$n"
done

echo "Installing nix"
su ubuntu <<'EOF'
  bash <(curl -L https://nixos.org/nix/install) --no-daemon
  chmod +x ~/.nix-profile/etc/profile.d/nix.sh
  ~/.nix-profile/etc/profile.d/nix.sh
EOF
chown -R ubuntu:ubuntu /nix

export DEVBOX_USE_VERSION="0.13.7"
export DEVBOX_USER="ubuntu"
export PATH="/home/${DEVBOX_USER}/.nix-profile/bin:$PATH"

echo "Installing devbox"
curl -L https://get.jetify.com/devbox | bash -s -- -f
chown -R ubuntu:ubuntu /usr/local/bin/devbox
chmod +x /usr/local/bin/devbox

runuser -u ubuntu renovate
