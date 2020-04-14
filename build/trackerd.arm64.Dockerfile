FROM golang:1.14.1-alpine as build-stage
RUN apk add --no-cache git build-base libc6-compat
ADD . /app/
WORKDIR /app
RUN make trackerd

# production stage
FROM nginx:stable-alpine as production-stage
RUN apk add --no-cache libc6-compat
RUN mkdir /app
COPY --from=build-stage /app/cmd/bin/arm64/linux/trackerd /app/trackerd
CMD ["/app/trackerd"]

