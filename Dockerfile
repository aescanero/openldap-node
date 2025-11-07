FROM docker.io/golang:1.23-alpine AS builder

LABEL org.opencontainers.image.authors="Alejandro Escanero Blanco <aescanero@disasterproject.com>"

USER 0

RUN apk --no-cache add ca-certificates && mkdir /data

WORKDIR /data/
RUN mkdir -p apiserver/dashboard/build
COPY cmd cmd
COPY config config
COPY go* .
COPY ldaputils ldaputils
COPY main.go .
COPY service service
COPY utils utils
COPY vendor vendor

COPY apiserver/*.go apiserver/.
COPY apiserver/dashboard/build apiserver/dashboard/build/.
#COPY go.sum .
#COPY app.go .

RUN go build -a -installsuffix cgo -o controller .

FROM docker.io/debian:stable-slim As server

LABEL org.opencontainers.image.authors="Alejandro Escanero Blanco <aescanero@disasterproject.com>"

RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install --no-install-recommends -y \
        slapd ldap-utils gettext-base procps ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    mv /etc/ldap /etc/openldap && \
    rm -f /var/lib/ldap/*

VOLUME [ "/etc/ldap" ]
VOLUME [ "/var/lib/ldap" ]

RUN chgrp -R 0 /var/lib/ldap && chmod -R g=u /var/lib/ldap && chmod u+x /var/lib/ldap && \
    chgrp -R 0 /etc/ldap && chmod -R g=u /etc/ldap && chmod u+x /etc/ldap
  
#USER 1001

WORKDIR /

EXPOSE 1389 1636

ENTRYPOINT ["/controller"]
CMD ["start"]

FROM server

LABEL org.opencontainers.image.authors="Alejandro Escanero Blanco <aescanero@disasterproject.com>"

COPY --from=builder /data/controller /.

WORKDIR /

EXPOSE 1389 1636

ENTRYPOINT ["/controller"]
CMD ["start"]
