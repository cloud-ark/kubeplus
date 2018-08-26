FROM fedora
ADD operator-manager /
#RUN mkdir /.helm && mkdir -p /.helm/repository && mkdir /.helm/repository/cache
#ADD repositories.yaml /.helm/repository
#ADD cloudark-helm-charts-index.yaml /.helm/repository/cache
ENTRYPOINT ["/operator-manager"]
