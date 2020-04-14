FROM alpine:3.11

ENV PACKAGES="gpsd"

RUN apk update && \
    apk add --no-cache ${PACKAGES}

EXPOSE 2947
ENTRYPOINT ["/bin/sh", "-c", "/sbin/syslogd -S -O - -n & exec /usr/sbin/gpsd -N -n -G ${*}","--"]

#CMD ["/bin/sh", "-c", "/sbin/syslogd -S -O - -n & exec /usr/sbin/gpsd -N -n -G ${*}","--"]
