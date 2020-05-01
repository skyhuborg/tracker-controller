FROM ubuntu:focal

ENV PACKAGES="sqlite3 x264 nano wget"

RUN DEBIAN_FRONTEND=noninteractive apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install --no-install-recommends --yes ${PACKAGES}

RUN wget https://dl.google.com/go/go1.13.6.linux-arm64.tar.gz && \
    tar -C /usr/local -xzf go1.13.6.linux-arm64.tar.gz

ENV GOPATH=/go/
ENV GO111MODULE=on
ENV PATH=${PATH}:/usr/local/go/bin

RUN mkdir /build
COPY go.mod /build/go.mod
COPY go.sum /build/go.sum
RUN cd /build && go mod download
ADD . /build/
RUN cd /build && \
    mkdir -p /uaptn/db && \
    mkdir -p /uaptn/etc && \
    cp /app/cmd/bin/arm64/linux/tracker-controller /skyhub/tracker-controller \
    ldd /uaptn/tracker-controller
WORKDIR /uaptn
CMD ["/uaptn/tracker-controller"]

# FROM golang:1.14.1-alpine as build-stage
# RUN apk add --no-cache git build-base
# ADD . /app/
# WORKDIR /app
# RUN echo $(pwd)
# RUN make tracker-controller

# # production stage
# FROM nginx:stable-alpine as production-stage
# RUN apk add --no-cache sqlite
# RUN mkdir -p /skyhub/data
# COPY --from=build-stage /app/cmd/bin/arm64/linux/tracker-controller /skyhub/tracker-controller
# EXPOSE 8088
# WORKDIR /skyhub
# CMD ["/skyhub/tracker-controller"]
