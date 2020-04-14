this works on x86_64
FROM golang:1.14.1-alpine
RUN apk add --no-cache python-dev linux-headers musl-dev
RUN apk add --no-cache git build-base sqlite pkgconfig curl alpine-sdk build-base apk-tools alpine-conf busybox fakeroot xorriso squashfs-tools cmake
RUN adduser --disabled-password build -G abuild
RUN echo "build ALL=(ALL) ALL" >> /etc/sudoers

ENV LD_LIBRARY_PATH=/usr/local/lib64
ENV PKG_CONFIG_PATH=/usr/local/lib/pkgconfig
ENV CMAKE_INSTALL_PREFIX=/usr/local
ENV WITH_QT=OFF
ENV WITH_GTK=OFF
ENV WITH_GSTREAMER=OFF

RUN echo "###### Build opencv #######" && \
    go get -u -d gocv.io/x/gocv && \
    cd $GOPATH/src/gocv.io/x/gocv && \
    make deps && make download && make build && make sudo_install && make clean

ADD . /build/    
RUN echo "###### Build tracker #######"; \
    find / -iname opencv4.pc; \
    mkdir /uaptn; \
    cd /build; \
    /build/scripts/create_trackerdb.sh; \
    cp /build/tracker.db /uaptn/data/tracker.db; \
    export PKG_CONFIG_PATH=/usr/local/lib64/pkgconfig; \
    export CMAKE_INSTALL_PREFIX=/usr/local; \
    echo $PKG_CONFIG_PATH; \
    make tracker; \
    cp /build/cmd/bin/amd64/linux/tracker /uaptn/tracker;

RUN echo "###### Clean the image #######" && \
    apk del ${DEV_DEPS} && \
    rm -rf /tmp/opencv && \
    rm -rf /var/cache/apk/*    

WORKDIR /uaptn
CMD ["/uaptn/tracker"]
