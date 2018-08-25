FROM fedora
ADD operator-deployer /
RUN mkdir /.helm && mkdir -p /.helm/repository && mkdir /.helm/repository/cache && mkdir -p /.helm/cache/archive && mkdir -p /.helm/cache/plugins
ADD repositories.yaml /.helm/repository
ADD cloudark-helm-charts-index.yaml /.helm/repository/cache
ENTRYPOINT ["/operator-deployer"]
