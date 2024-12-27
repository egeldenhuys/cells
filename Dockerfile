# Handle certificate and download in a distinct stage to reduce image size
FROM docker.io/golang:1.23 as builder
WORKDIR /src
COPY . .
RUN make build

FROM docker.io/alpine as certs
ARG version

RUN apk update && apk add ca-certificates
ENV CELLS_VERSION ${version}

WORKDIR /pydio
COPY --from=builder /src/cells .
RUN wget --output-document=jq "https://download.pydio.com/pub/linux/tools/jq-linux64-v1.6"
RUN chmod +x /pydio/cells /pydio/jq 

# Create the target image
FROM docker.io/busybox:glibc
ARG version

# Add necessary files
COPY ./tools/docker/images/cells/docker-entrypoint.sh /opt/pydio/bin/docker-entrypoint.sh
COPY ./tools/docker/images/cells/libdl.so.2 /opt/pydio/bin/libdl.so.2
COPY --from=certs /pydio/jq /bin/jq
COPY --from=certs /etc/ssl/certs /etc/ssl/certs
COPY --from=certs /pydio/cells /opt/pydio/bin/cells

ENV CADDYPATH /var/cells/certs 
ENV CELLS_WORKING_DIR /var/cells
WORKDIR $CELLS_WORKING_DIR

# Final configuration
RUN ln -s /opt/pydio/bin/cells /bin/cells \
    && ln -s /opt/pydio/bin/libdl.so.2 /lib64/libdl.so.2 \
    && ln -s /opt/pydio/bin/docker-entrypoint.sh /bin/docker-entrypoint.sh \
    && chmod +x /opt/pydio/bin/docker-entrypoint.sh \
    && echo "Pydio Cells Home Docker Image" > /opt/pydio/package.info \
    && echo "  A ready-to-go Docker image based on BusyBox to configure and launch Cells in no time." >> /opt/pydio/package.info \
    && echo "  Generated on $(date) with docker build script from version ${version}" >> /opt/pydio/package.info

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["cells", "start"]
