Hermes can be configured via a custom JSON configuration, and the path can be passed over via the `CONFIG_FILE_PATH` environment variable.
The current JSON schema to be adhered to, can be found at [**c16a/hermes:/config/config.go**](https://github.com/c16a/hermes/blob/master/config/config.go)

When running on Docker or Kubernetes, this file should be mounted as a volume.
