FROM golang:1.14.1-alpine as build-stage
RUN apk add --no-cache git build-base libc6-compat
ADD . /app/
WORKDIR /app
RUN echo $(pwd)
RUN make portald

# production stage
FROM nginx:stable-alpine as production-stage
RUN apk add --no-cache libc6-compat
RUN mkdir /uaptn
COPY --from=build-stage /app/cmd/bin/amd64/linux/portald /uaptn/portald
CMD ["/uaptn/portald"]