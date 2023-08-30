FROM alpine:latest
LABEL authors="itaylevy"
RUN apk --no-cache add curl
ENV ROLLER_RELEASE_TAG=v0.1.12-beta-rc2
RUN curl -L -o install.sh https://dymensionxyz.github.io/roller/install.sh
RUN chmod +x install.sh
RUN ./install.sh
