FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-dockerhub"]
COPY baton-dockerhub /