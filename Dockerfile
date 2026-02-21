# You can easily repackage Keycloak with a latest version by using this arg during the docker build
ARG KEYCLOAK_VERSION=26.5
ARG KC_SERVER_HOSTNAME=keycloak.kong.runlocal.dev

# See https://www.keycloak.org/server/containers
FROM quay.io/keycloak/keycloak:${KEYCLOAK_VERSION} AS builder
ARG KC_SERVER_HOSTNAME

WORKDIR /opt/keycloak

ENV KC_HOSTNAME=${KC_SERVER_HOSTNAME}
ENV KC_HTTP_ENABLED=true
ENV KC_PROXY_HEADERS=xforwarded

# Keycloak v26.5 comes with preview of capabilities on grants like kubernetes-service-accounts, spiffe, and jwt-authorization-grant
RUN /opt/keycloak/bin/kc.sh build --features=preview


# From a java image
FROM quay.io/keycloak/keycloak:${KEYCLOAK_VERSION}
ARG KC_SERVER_HOSTNAME

COPY --from=builder /opt/keycloak/ /opt/keycloak/
WORKDIR /opt/keycloak/bin

ENV KC_HOSTNAME=${KC_SERVER_HOSTNAME}
ENV KC_HTTP_ENABLED=true
ENV KC_PROXY_HEADERS=xforwarded

# open a new shell with bash for easier scripting
ENTRYPOINT ["/bin/bash"]
# CMD ["/bin/bash" , "-c", "sleep infinity"]

