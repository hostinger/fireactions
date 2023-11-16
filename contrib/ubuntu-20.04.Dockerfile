FROM golang:1.20 AS fireactions-agent

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

ENV GO111MODULE=on \
    CGO_ENABLED=0  \
    GOOS=linux     \
    GOARCH=amd64

RUN go build -v -o fireactions-agent ./cmd/fireactions-agent

FROM ubuntu:20.04

ARG TARGETPLATFORM
ARG RUNNER_VERSION

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

RUN export ARCH=$(echo ${TARGETPLATFORM} | cut -d / -f2)                                                                                                    \
    && if [ "$ARCH" = "amd64" ] || [ "$ARCH" = "x86_64" ] || [ "$ARCH" = "i386" ]; then export ARCH=x64 ; fi                                                \
    && mkdir -p /opt/runner && cd /opt/runner                                                                                                               \
    && curl -fLo runner.tar.gz https://github.com/actions/runner/releases/download/v${RUNNER_VERSION}/actions-runner-linux-${ARCH}-${RUNNER_VERSION}.tar.gz \
    && tar xzf ./runner.tar.gz && rm -rf runner.tar.gz                                                                                                      \
    && ./bin/installdependencies.sh                                                                                                                         \
    && chown -R runner:docker /opt/runner

RUN install -m 0755 -d /etc/apt/keyrings                                                                            \ 
    && curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg      \
    && chmod a+r /etc/apt/keyrings/docker.gpg                                                                       \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
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

COPY --from=fireactions-agent /app/fireactions-agent /usr/bin/fireactions-agent
COPY contrib/overlay/etc /etc

RUN systemctl enable fireactions-agent.service
