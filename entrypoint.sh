#!/bin/sh

# Perform environment variable substitution on config.template.yaml
envsubst < /app/configs/config.template.yaml > /app/configs/config.yaml
exec "$@"
