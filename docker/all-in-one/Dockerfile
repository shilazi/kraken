# syntax=docker/dockerfile:1.2
FROM docker.io/library/golang:1.14 as builder

ENV GO111MODULE=on

RUN set -x \
    && DEBIAN_FRONTEND=noninteractive apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y \
         --no-install-recommends rsync rename

ADD go.mod go.sum /go/src/github.com/uber/kraken/

WORKDIR /go/src/github.com/uber/kraken/

RUN --mount=type=cache,sharing=locked,id=gomod,target=/go/pkg/mod/cache \
      go mod download

ADD . /go/src/github.com/uber/kraken/

ARG LINUX_BINS="agent/agent build-index/build-index origin/origin proxy/proxy tools/bin/testfs/testfs tracker/tracker"

RUN set -x \
    && mkdir /tmp/bin \
    && for lb in ${LINUX_BINS}; do \
        go build -ldflags "-s -w" -o ./${lb} ./$(dirname ${lb})/ \
        && rsync -a ./${lb} /tmp/bin/ \
    ; done \
    && cd /tmp/bin \
    && rename 's/(^[^k])/kraken-$1/' *

# ---------- 8< ----------

FROM docker.io/library/debian:12

RUN set -x \
    && DEBIAN_FRONTEND=noninteractive apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y \
        --no-install-recommends \
        ca-certificates curl sqlite3 nginx procps

RUN set -x \
    && mkdir -p -m 755 /var/kraken/run \
    && mkdir -p -m 755 /var/kraken/log \
    && mkdir -p -m 755 /var/kraken/cache

COPY --from=builder /tmp/bin/ /usr/bin/

WORKDIR /etc/kraken

ADD test/tls/ /etc/kraken/tls/
ADD config/ /etc/kraken/config/
ADD nginx/config/ /etc/kraken/nginx/config/
