FROM debian:bullseye as build-stage
ARG GOLANG_INSTALL=https://go.dev/dl/go1.20.6.linux-amd64.tar.gz

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y build-essential libudev-dev ca-certificates curl

RUN update-ca-certificates

RUN curl -L ${GOLANG_INSTALL} | tar -C /usr/local -xzf -
ENV PATH="${PATH}:/usr/local/go/bin"

RUN adduser --disabled-password --gecos '' spacemesh 
COPY --chown=spacemesh . /home/spacemesh 
USER spacemesh
WORKDIR /home/spacemesh

RUN make build

FROM scratch as export-stage
COPY --from=build-stage /home/spacemesh/smcli /
