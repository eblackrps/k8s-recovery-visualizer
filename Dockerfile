FROM --platform=$BUILDPLATFORM golang:1.25 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags "-s -w -X 'k8s-recovery-visualizer/internal/model.Version=${VERSION}' -X 'k8s-recovery-visualizer/internal/model.BuildDate=${BUILD_DATE}'" \
    -o /out/scan ./cmd/scan

FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/scan /usr/local/bin/scan
ENTRYPOINT ["/usr/local/bin/scan"]
