FROM mcr.microsoft.com/devcontainers/go:1.25-bookworm

# [Optional] Uncomment this section to install additional OS packages.
RUN rm -f /etc/apt/sources.list.d/yarn.list && apt-get update \
    && apt-get -y install --no-install-recommends postgresql-client \
    && rm -rf /var/lib/apt/lists/*

# [Optional] Uncomment the next lines to use go get to install anything else you need
# USER vscode
RUN go install github.com/a-h/templ/cmd/templ@latest

COPY --from=1password/op:2 /usr/local/bin/op /usr/local/bin/op

# [Optional] Uncomment this line to install global node packages.
# RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && npm install -g <your-package-here>" 2>&1