FROM --platform=$TARGETPLATFORM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates                                              \
    && apt-get autoremove -y                                     \
    && apt-get clean                                             \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*             \
    && groupadd -g 1000 fireactions                              \
    && useradd -u 1000 -g fireactions -s /bin/sh -m fireactions

COPY fireactions /usr/bin/fireactions
RUN chmod +x /usr/bin/fireactions

EXPOSE 8080

COPY entrypoint.sh /usr/bin/entrypoint.sh

ENTRYPOINT ["/usr/bin/entrypoint.sh"]
