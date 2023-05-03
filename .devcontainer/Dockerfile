FROM mcr.microsoft.com/devcontainers/go:0-1.20

# [Optional] Uncomment this section to install additional OS packages.
# RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
#     && apt-get -y install --no-install-recommends <your-package-list-here>

ARG DBMATE_VERSION=v2.2.0
ARG SQLC_VERSION=1.18.0
ARG AIR_VERSION=v1.43.0

RUN curl -fsSL -o /usr/local/bin/dbmate https://github.com/amacneil/dbmate/releases/download/${DBMATE_VERSION}/dbmate-linux-amd64 \
    && sudo chmod +x /usr/local/bin/dbmate
RUN curl -fsSL https://github.com/kyleconroy/sqlc/releases/download/v${SQLC_VERSION}/sqlc_${SQLC_VERSION}_linux_amd64.tar.gz -o /tmp/sqlc.tgz \
    && tar xzf /tmp/sqlc.tgz --directory /usr/local/bin/

# [Optional] Uncomment the next lines to use go get to install anything else you need
USER vscode
RUN go install github.com/cosmtrek/air@${AIR_VERSION}
USER root

# [Optional] Uncomment this line to install global node packages.
# RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && npm install -g <your-package-here>" 2>&1
