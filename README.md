# Nocloud Chat Service
## Building Proto
In the project root run:
```sh
docker run -it \
  -v $(pwd)/pkg:/go/src/github.com/slntopp/nocloud/pkg \
  ghcr.io/slntopp/nocloud/buf:latest
```
