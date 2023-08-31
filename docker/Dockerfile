FROM ubuntu:latest
LABEL authors="itaylevy"
RUN apt-get update && apt-get install -y curl
RUN apt-get update && apt-get install -y sudo
# install jq
RUN apt-get update && apt-get install -y jq
ENV ROLLER_RELEASE_TAG=v0.1.12-beta-rc2
RUN curl -L -o install.sh https://dymensionxyz.github.io/roller/install.sh
RUN chmod +x install.sh
RUN ./install.sh
WORKDIR /app
COPY . .
# Make the entrypoint script executable
RUN chmod +x /app/entrypoint.sh

# Define the script we want run once the container boots up
ENTRYPOINT ["/app/entrypoint.sh"]

