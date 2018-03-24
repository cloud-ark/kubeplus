Grapher
--------
Generates graphical output for kubectl connections plugin


Development
-----------
Inside kubediscovery:
- go build .
- ./kubediscovery connections Service wordpress-wp1 default -o json --ignore=ServiceAccount:default,Namespace:default > reldetails.json

Inside grapher:
- cp ../kubediscovery/reldetails.json .
- docker build -t grapher-reldetails .
- docker run -v `pwd`:/root grapher-reldetails:latest reldetails.json /root


