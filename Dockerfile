FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-buildkite"]
COPY baton-buildkite /