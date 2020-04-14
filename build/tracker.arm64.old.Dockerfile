FROM golang:1.14.1-alpine

ENV BUILD="ca-certificates libjpeg-turbo libpng libwebp \
         libwebp-dev tiff openblas libgphoto2 \
         sqlite"

ENV DEV="clang clang-dev cmake pkgconf git build-base musl-dev alpine-sdk make \
         openblas-dev gcc g++ libc-dev linux-headers \
         libgphoto2-dev libjpeg-turbo-dev libpng-dev \
         tiff-dev"

ARG OPENCV_VERSION="4.2.0"
ENV OPENCV_VERSION $OPENCV_VERSION

ENV LD_LIBRARY_PATH=/usr/local/lib64 \
    PKG_CONFIG_PATH=/usr/local/lib64/pkgconfig \
    CMAKE_INSTALL_PREFIX=/usr/local \
    # CFLAGS="mfloat-abi=hard -mfpu=vfpv3" \
    # CXXFLAGS="mfloat-abi=hard -mfpu=vfpv3"

RUN echo "###### Install opencv #######" && \
    apk update && \
    apk add --no-cache ${BUILD} ${DEV} && \
    go get -u -d gocv.io/x/gocv && \
    cd $GOPATH/src/gocv.io/x/gocv && \
    make deps && make download && make -j build && make sudo_install && make clean && \
    rm -rf /tmp/opencv && \
    apk del ${DEV_DEPS} && \
    rm -rf /var/cache/apk/* 

ADD . /build/    
RUN echo "###### Build tracker #######" && \
    cd /build && \
    mkdir /app && \
    /build/scripts/create_trackerdb.sh && \
    cp tracker.db /app/tracker.db && \
    make tracker && \
    cp /build/cmd/bin/arm64/linux/tracker /app/tracker && \
    rm -rf /build && \
    rm -rf /go/pkg/mod && \
    ldd /app/tracker
    
WORKDIR /app

CMD ["/app/tracker"]
