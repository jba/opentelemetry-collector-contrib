# HTTP Check Receiver

<!-- status autogenerated section -->
| Status        |           |
| ------------- |-----------|
| Stability     | [development]: metrics   |
| Distributions | [contrib], [sumo] |
| Issues        | [![Open issues](https://img.shields.io/github/issues-search/open-telemetry/opentelemetry-collector-contrib?query=is%3Aissue%20is%3Aopen%20label%3Areceiver%2Fhttpcheck%20&label=open&color=orange&logo=opentelemetry)](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues?q=is%3Aopen+is%3Aissue+label%3Areceiver%2Fhttpcheck) [![Closed issues](https://img.shields.io/github/issues-search/open-telemetry/opentelemetry-collector-contrib?query=is%3Aissue%20is%3Aclosed%20label%3Areceiver%2Fhttpcheck%20&label=closed&color=blue&logo=opentelemetry)](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues?q=is%3Aclosed+is%3Aissue+label%3Areceiver%2Fhttpcheck) |
| [Code Owners](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/CONTRIBUTING.md#becoming-a-code-owner)    | [@codeboten](https://www.github.com/codeboten) |

[development]: https://github.com/open-telemetry/opentelemetry-collector#development
[contrib]: https://github.com/open-telemetry/opentelemetry-collector-releases/tree/main/distributions/otelcol-contrib
[sumo]: https://github.com/SumoLogic/sumologic-otel-collector
<!-- end autogenerated section -->

The HTTP Check Receiver can be used for synthethic checks against HTTP endpoints. This receiver will make a request to the specified `endpoint` using the
configured `method`. This scraper generates a metric with a label for each HTTP response status class with a value of `1` if the status code matches the
class. For example, the following metrics will be generated if the endpoint returned a `200`:

```
httpcheck.status{http.status_class:1xx, http.status_code:200,...} = 0
httpcheck.status{http.status_class:2xx, http.status_code:200,...} = 1
httpcheck.status{http.status_class:3xx, http.status_code:200,...} = 0
httpcheck.status{http.status_class:4xx, http.status_code:200,...} = 0
httpcheck.status{http.status_class:5xx, http.status_code:200,...} = 0
```

## Configuration

The following configuration settings are required:

- `endpoint`: The URL of the endpoint to be monitored.

The following configuration settings are optional:

- `method` (default: `GET`): The method used to call the endpoint.
- `collection_interval` (default = `60s`): This receiver collects metrics on an interval. Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`.
- `initial_delay` (default = `1s`): defines how long this receiver waits before starting.

### Example Configuration

```yaml
receivers:
  httpcheck:
    targets:
      - endpoint: http://endpoint:80
        method: GET
      - endpoint: http://localhost:8080/health
        method: GET
    collection_interval: 10s
```

## Metrics

Details about the metrics produced by this receiver can be found in [documentation.md](./documentation.md)
