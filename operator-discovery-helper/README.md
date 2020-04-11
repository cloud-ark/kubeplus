Operator Discovery Helper
--------------------------

Update versions.txt and then use ./build-artifact.sh to build deployable artifact
(Docker image).

The module dependencies are captured in Gopkg.lock, Gopkg.toml and go.mod.
go.mod was created from Gopkg file.
client-go dependency in go.mod has been updated from that in Gopkg.lock.
The code in main.go has also been updated to use the newer version of client-go.

Use Go 1.13 to build Operator Discovery Helper.

setgopaths.sh is provided to set the GOROOT and update PATH.

Download Golang 1.13 and update setgopath.sh to reflect the correct path.

vendor directory contains dependencies created using Go version 1.10.2.
It has been left here for historical purposes.


