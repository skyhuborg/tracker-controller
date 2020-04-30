FROM golang:1.14.1-alpine as build-stage
RUN apk add --no-cache git build-base
ADD . /app/
WORKDIR /app
RUN echo $(pwd)
RUN make tracker-controller

# production stage
FROM nginx:stable-alpine as production-stage
RUN apk add --no-cache sqlite
RUN mkdir -p /skyhub/data
COPY --from=build-stage /app/cmd/bin/arm64/linux/tracker-controller /skyhub/tracker-controller
EXPOSE 8088
WORKDIR /skyhub
CMD ["/skyhub/tracker-controller"]
