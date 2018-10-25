FROM fedora
COPY operator-deployer /
RUN mkdir /.helm && mkdir -p /.helm/repository && mkdir /.helm/repository/cache && mkdir -p /.helm/cache/archive && mkdir -p /.helm/cache/plugins
COPY repositories.yaml /.helm/repository/
COPY cloudark-helm-charts-index.yaml /.helm/repository/cache/
ENTRYPOINT ["/operator-deployer"]
