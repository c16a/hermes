Hermes uses a multi stage docker build for hermetic builds, 
while creating a minimal image. Hence, please ensure you use Docker v17.05 or newer.

```shell
git clone https://github.com/c16a/hermes.git
cd hermes
docker build -t hermes-app .
```

### Running the image
```shell
docker run -p 4000:4000 -v $pwd/config.json:/app/config.json hermes-app
```

The above example assumes that the TCP server has been 
configured to listen on port 4000. 
In case that is configured to another port, 
please configure the docker exposed port accordingly.

The [Configuration](configuration.md) section has more details on which attributes of the broker can be configured.
