FROM ghcr.io/berriai/litellm:v1.81.3-stable

COPY config.yaml /app/config.yaml
EXPOSE 4000
ENTRYPOINT ["litellm", "--config", "/app/config.yaml", "--port", "4000"]
