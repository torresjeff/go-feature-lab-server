# Feature Lab server
Backend for Go Feature Lab, a feature flag solution for Go.

Build the image:
```shell
$ docker image build -t torresjeff/go-feature-lab-server .
```

Run the container
```shell
$ docker container run --name featurelab -d -p 3000:3000 torresjeff/go-feature-lab-server
```