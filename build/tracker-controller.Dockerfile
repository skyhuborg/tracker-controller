FROM golang:1.14.1-alpine as build-stage
RUN apk add --no-cache git build-base
ADD . /app/
WORKDIR /app
RUN echo $(pwd)
RUN make tracker-controller

# production stage
FROM nginx:stable-alpine as production-stage
RUN apk add --no-cache sqlite
RUN mkdir -p /uaptn/data
COPY --from=build-stage /app/cmd/bin/arm64/linux/tracker-controller /uaptn/tracker-controller
EXPOSE 8088
WORKDIR /uaptn
CMD ["/uaptn/tracker-ui-backend"]
