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

ENV LD_LIBRARY_PATH=/usr/local/lib \
    PKG_CONFIG_PATH=/usr/local/lib/pkgconfig \
    CMAKE_INSTALL_PREFIX=/usr/local

RUN echo "###### Build and Install opencv #######" && \
    apk update && \
    apk add --no-cache ${BUILD} ${DEV} && \
    curl -Lo opencv.zip https://github.com/opencv/opencv/archive/${OPENCV_VERSION}.zip && \
    unzip -q opencv.zip && \
    curl -Lo opencv_contrib.zip https://github.com/opencv/opencv_contrib/archive/${OPENCV_VERSION}.zip && \
    unzip -q opencv_contrib.zip && \
    rm opencv.zip opencv_contrib.zip && \
    cd opencv-${OPENCV_VERSION} && \
    mkdir build && cd build && \
    cmake -D CMAKE_BUILD_TYPE=RELEASE \
        -D CMAKE_INSTALL_PREFIX=/usr/local \
        -D OPENCV_EXTRA_MODULES_PATH=../../opencv_contrib-${OPENCV_VERSION}/modules \
        -D WITH_QT=OFF \
        -D WITH_GTK=OFF \
        -D WITH_GSTREAMER=OFF \
        -D WITH_OPENCL=OFF \
        -D BUILD_DOCS=OFF \
        -D BUILD_EXAMPLES=OFF \
        -D BUILD_TESTS=OFF \
        -D BUILD_PROTOBUF=OFF \
        -D BUILD_PERF_TESTS=OFF \
        -D BUILD_opencv_java=OFF \
        -D BUILD_opencv_python=NO \
        -D BUILD_opencv_python2=NO \
        -D BUILD_opencv_python3=NO \
        -D ENABLE_NEON=ON \
        -D ENABLE_VFPV3=ON \
        -D WITH_JASPER=OFF \
        -D OPENCV_ENABLE_NONFREE=ON \
        -D ENABLE_PRECOMPILED_HEADERS=OFF \
        -D OPENCV_EXTRA_EXE_LINKER_FLAGS=-latomic \
        -D OPENCV_GENERATE_PKGCONFIG=ON ..
RUN cd opencv-${OPENCV_VERSION} && make preinstall
RUN cd opencv-${OPENCV_VERSION} && make install
RUN cd opencv-${OPENCV_VERSION} && ldconfig
RUN cd / && rm -rf opencv*

# RUN echo "###### Install opencv #######" && \
#     apk update && \
#     apk add --no-cache ${BUILD} ${DEV} && \
#     go get -u -d gocv.io/x/gocv && \
#     cd $GOPATH/src/gocv.io/x/gocv && \
#     make deps && make download && make -j build && make sudo_install && make clean && \
#     rm -rf /tmp/opencv && \
#     apk del ${DEV_DEPS} && \
#     rm -rf /var/cache/apk/* 

ADD . /build/    
RUN echo "###### Build tracker #######" && \
    cd /build && \
    mkdir /uaptn/db && \
    /build/scripts/create_trackerdb.sh && \
    cp tracker.db /uaptn/db/tracker.db && \
    make tracker && \
    cp /build/cmd/bin/arm64/linux/tracker /uaptn/tracker && \
    rm -rf /build && \
    rm -rf /go/pkg/mod && \
    ldd /uaptn/tracker


    
WORKDIR /uaptn

CMD ["/uaptn/tracker"]
