FROM registry.gitlab.com/uaptn/gocv-alpine:4.2.0-buildstage as build-stage
FROM golang:1.14.1-alpine3.11

COPY --from=build-stage /usr/local/lib64 /usr/local/lib64
COPY --from=build-stage /usr/local/lib64/pkgconfig/opencv4.pc /usr/local/lib64/pkgconfig/opencv4.pc
COPY --from=build-stage /usr/local/include/opencv4/opencv2 /usr/local/include/opencv4/opencv2

ENV PACKAGES="ca-certificates libjpeg-turbo libpng libwebp \
          libwebp-dev tiff openblas libgphoto2 \
          sqlite clang clang-dev cmake pkgconf git build-base musl-dev alpine-sdk make \
          openblas-dev gcc g++ libc-dev linux-headers \
          libgphoto2-dev libjpeg-turbo-dev libpng-dev \
          tiff-dev sqlite make git gcc pkgconfig build-base" \
    LD_LIBRARY_PATH=/usr/local/lib64 \
    PKG_CONFIG_PATH=/usr/local/lib64/pkgconfig          
RUN apk update && \
    apk add --no-cache ${PACKAGES}
ADD . /build/
RUN cd /build && \
    mkdir -p /uaptn/db && \
    mkdir -p /uaptn/etc && \
    scripts/create_trackerdb.sh && \
    cp tracker.db /uaptn/db/tracker.db && \
    make tracker && \
    cp /build/cmd/bin/arm64/linux/tracker /uaptn/tracker && \
    ldd /uaptn/tracker
WORKDIR /uaptn
CMD ["/uaptn/tracker"]