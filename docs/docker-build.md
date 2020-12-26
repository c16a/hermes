Hermes uses a multi stage docker build for hermetic builds, while creating a minimal image. Hence, please ensure you use
Docker v17.05 or newer.

```shell
git clone https://github.com/c16a/hermes.git
cd hermes
docker build -t hermes-app .
```

### Running the image

```shell
docker run -p 4000:4000 -v $pwd/config.json:/app/config.json hermes-app
```

The above example assumes that the TCP server has been configured to listen on port 4000. In case that is configured to
another port, please configure the docker exposed port accordingly.

#### SELinux policies

When using Docker on a host with SELinux enabled, the container is denied access to certain parts of host file system
unless it is run in privileged mode. To resolve this, you can use a named volume

```shell
# Create a docker volume and map it to /tmp/hermes on the host
docker volume create --driver local --opt type=none --opt device=/tmp/hermes --opt o=bind hermes_volume

# Ensure /tmp/hermes/config.json has the required broker configuration
# Use the above created hermes_volume to mount the config file into the container
docker run -p 4000:4000 -e CONFIG_FILE_PATH=/tmp/hermes/config.json --mount source=hermes_volume,target=/tmp/hermes hermes
```

Please note that however, you place your `config.json` in the `/tmp` directory, SELinux does not restrict you access
when you use a direct volume mapping.

```shell
# This won't work with SELinux enabled
docker run -p 4000:4000 -e CONFIG_FILE_PATH=/tmp/hermes/config.json -v /home/user/config.json:/tmp/hermes/config.json hermes

# This will work
docker run -p 4000:4000 -e CONFIG_FILE_PATH=/tmp/hermes/config.json -v /tmp/hermes/config.json:/tmp/hermes/config.json hermes
```

The [Configuration](configuration.md) section has more details on which attributes of the broker can be configured.

### Running in Compose mode

Create the named volume `hermes_volume`.

```shell
# Create a docker volume and map it to /tmp/hermes on the host
docker volume create --driver local --opt type=none --opt device=/tmp/hermes --opt o=bind hermes_volume
```

Reference the named volume for the service

```yaml
version: "3.9"
services:
  broker:
    build:
      context: .
    environment:
      CONFIG_FILE_PATH: "/tmp/hermes/config.json"
    volumes:
      - hermes_volume:/tmp/hermes
    ports:
      - 4000:4000
      - 5000:5000
volumes:
  hermes_volume:
    external: true
```
