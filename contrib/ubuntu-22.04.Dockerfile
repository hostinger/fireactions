FROM ubuntu:22.04

ARG ARCH
ARG RUNNER_VERSION
ARG AGENT_VERSION

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update -y                              \
  && apt-get install -y software-properties-common \
  && add-apt-repository -y ppa:git-core/ppa        \
  && apt-get update -y                             \
  && apt-get install -y --no-install-recommends    \
    curl                                           \
    dbus                                           \
    kmod                                           \
    iproute2                                       \
    iputils-ping                                   \
    iptables-persistent                            \
    iptables                                       \
    cmake                                          \
    build-essential                                \
    pkg-config                                     \
    libssl-dev                                     \
    net-tools                                      \
    openssh-server                                 \
    haveged                                        \
    sudo                                           \
    systemd                                        \
    udev                                           \
    unzip                                          \
    ca-certificates                                \
    jq                                             \
    zip                                            \
    gnupg                                          \
    git-lfs                                        \
    git                                            \
    vim-tiny                                       \
    wget

RUN systemctl enable haveged.service

RUN update-alternatives --set iptables /usr/sbin/iptables-legacy && \
    update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy

RUN adduser --disabled-password --gecos "" --uid 1001 runner  \
    && groupadd docker --gid 121                              \
    && usermod -aG docker runner                              \
    && usermod -aG sudo runner                                \
    && echo "%sudo   ALL=(ALL:ALL) NOPASSWD:ALL" > /etc/sudoers

RUN case "$ARCH" in amd64|x86_64|i386) export RUNNER_ARCH="x64";; arm64) export RUNNER_ARCH="arm64";; esac                                                         \
    && mkdir -p /opt/runner /opt/hostedtoolcache && cd /opt/runner                                                                                                 \
    && curl -fLo runner.tar.gz https://github.com/actions/runner/releases/download/v${RUNNER_VERSION}/actions-runner-linux-${RUNNER_ARCH}-${RUNNER_VERSION}.tar.gz \
    && tar xzf ./runner.tar.gz && rm -rf runner.tar.gz                                                                                                             \
    && ./bin/installdependencies.sh                                                                                                                                \
    && chown -R runner:docker /opt/runner /opt/hostedtoolcache                                                                                                     \
    && chmod -R 777 /opt/hostedtoolcache

RUN install -m 0755 -d /etc/apt/keyrings                                                                            \
    && curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg      \
    && chmod a+r /etc/apt/keyrings/docker.gpg                                                                       \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg]                         \
            https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo $VERSION_CODENAME) stable"         \
        | tee /etc/apt/sources.list.d/docker.list > /dev/null                                                       \
    && apt-get update && apt-get install --no-install-recommends -y                                                 \
        docker-ce                                                                                                   \
        docker-ce-cli                                                                                               \
        containerd.io                                                                                               \
        docker-buildx-plugin                                                                                        \
        docker-compose-plugin                                                                                       \
    && apt-get clean && rm -rf /var/lib/apt/lists/*                                                                 \
    && systemctl enable docker.service

RUN echo 'root:root' | chpasswd                                                                   \
    && sed -i -e 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config \
    && sed -i -e 's/^AcceptEnv LANG LC_\*$/#AcceptEnv LANG LC_*/'            /etc/ssh/sshd_config

RUN echo "" > /etc/machine-id && echo "" > /var/lib/dbus/machine-id

COPY fireactions-agent-${AGENT_VERSION}-linux-${ARCH} /tmp
COPY contrib/overlay/etc /etc

RUN mv /tmp/fireactions-agent-${AGENT_VERSION}-linux-${ARCH} /usr/bin/fireactions-agent && \
    chmod +x /usr/bin/fireactions-agent

RUN systemctl enable fireactions-agent.service
