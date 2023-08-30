FROM ubuntu:latest
LABEL authors="itaylevy"
RUN apt-get update && apt-get install -y curl
RUN export ROLLER_RELEASE_TAG="v0.1.12-beta-rc2" && curl -L https://dymensionxyz.github.io/roller/install.sh | bash
