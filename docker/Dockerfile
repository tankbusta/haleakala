FROM alpine:latest

# Add s6-overlay
ENV S6_OVERLAY_VERSION=v1.21.4.0

RUN apk add --update --no-cache bind-tools curl && \
    curl -sSL https://github.com/just-containers/s6-overlay/releases/download/${S6_OVERLAY_VERSION}/s6-overlay-amd64.tar.gz \
    | tar xfz - -C /

ADD root /

COPY haleakala /usr/local/bin/haleakala
RUN chmod +x /usr/local/bin/haleakala
ENV PATH /usr/local/bin:$PATH

WORKDIR /opt/haleakala

ENTRYPOINT ["/init"]
