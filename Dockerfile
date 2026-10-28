# LLD Prep: one container with the server, Go (for Go exercises) and a JDK
# (for Java exercises). Built for Render, works on any Docker host.

FROM golang:1.23-bookworm AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/lldprep .

FROM golang:1.23-bookworm
RUN apt-get update \
 && apt-get install -y --no-install-recommends openjdk-17-jdk-headless \
 && rm -rf /var/lib/apt/lists/* \
 && useradd --create-home --uid 10001 app \
 && mkdir -p /data && chown app:app /data

COPY --from=build /out/lldprep /usr/local/bin/lldprep
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

# Warm the Go build cache as the runtime user so the first test run is fast.
USER app
RUN cd /tmp && lldprep -verify -only strategy-discounts >/dev/null
USER root

ENV PORT=8080
EXPOSE 8080
VOLUME /data
ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["lldprep", "-data", "/data"]
