# radius-exporter

A prometheus exporter similar to blackbox_exporter but for RADIUS servers.


## Usage 
```
Usage of ./radius-exporter:
  -config string
        Path to the config file. (default "config.yml")
  -log.level string
        The log level. (default "info")
  -web.disable-exporter-metrics
        Exclude metrics about the exporter itself
  -web.listen-address string
        Address on which to expose metrics and web interface. (default ":9881")
  -web.tlemetry-path string
        Path under which to expose metrics (default "/metrics")
```

## Configuration

```yaml
---
modules:
  module_name:
    # Required
    username: radius-test
    # Required
    password: radius-test
    # RADIUS shared secret. Required.
    secret: radius-test
    # At least one of nas_id or nas_ip should be configured to comply with RFC2865 (4.1)
    nas_id: prometheus
    nas_ip: 192.0.2.1
```
