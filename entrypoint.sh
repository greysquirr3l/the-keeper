#!/bin/sh

# Perform environment variable substitution on config.template.yaml
envsubst < ./configs/config.template.yaml > ./configs/config.yaml

# Execute the main application
exec ./the-keeper
