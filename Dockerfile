# Build
FROM golang:alpine AS build

RUN apk add --no-cache -U build-base git make

RUN mkdir -p /src

WORKDIR /src

# Copy Makefile
COPY Makefile ./

# Copy go.mod and go.sum and install and cache dependencies
COPY go.mod .
COPY go.sum .

# Install deps
RUN make deps
RUN go mod download

# Copy static assets
COPY ./ui/css/* ./ui/css/
COPY ./ui/img/* ./ui/img/
COPY ./ui/js/* ./ui/js/
COPY ./ui/js/actions/* ./ui/js/actions/
COPY ./ui/js/components/* ./ui/js/components/
COPY ./ui/js/utils/* ./ui/js/utils/
COPY ./ui/vendor/* ./ui/vendor/
COPY ./ui/* ./ui/

# Copy sources
COPY *.go ./
COPY ./blob/*.go ./blob/
COPY ./store/*.go ./store/

# Version/Commit (there there is no .git in Docker build context)
# NOTE: This is fairly low down in the Dockerfile instructions so
#       we don't break the Docker build cache just be changing
#       unrelated files that actually haven't changed but caused the
#       COMMIT value to change.
ARG VERSION="0.0.0"
ARG COMMIT="HEAD"

# Build fbox binary
RUN make fbox VERSION=$VERSION COMMIT=$COMMIT

# Runtime
FROM alpine:latest

RUN apk --no-cache -U add ca-certificates tzdata

WORKDIR /
VOLUME /data

# force cgo resolver
ENV GODEBUG=netdns=cgo

COPY --from=build /src/fbox /fbox

ENTRYPOINT ["/fbox"]
CMD [""]
