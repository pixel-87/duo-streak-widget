#!/usr/bin/env bash

set -euo pipefail

# Script to build, load, tag, and push the Duo Streak Widget image to Google Container Registry

echo "Building Nix image..."
nix build .#image

echo "Loading image into Docker..."
docker load < result

echo "Tagging image for GCR..."
docker tag duo-streak-widget:latest gcr.io/duo-streak-widget/duo-streak-widget:latest

echo "Pushing to GCR..."
echo "Note: Make sure you're authenticated with 'gcloud auth configure-docker' or similar."
docker push gcr.io/duo-streak-widget/duo-streak-widget:latest

echo "Image pushed successfully! You can now run 'terraform apply' to deploy."