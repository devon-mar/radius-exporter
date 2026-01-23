module github.com/devon-mar/radius-exporter

go 1.19

require (
	github.com/prometheus/client_golang v1.19.1
	gopkg.in/yaml.v3 v3.0.1
	layeh.com/radius v0.0.0-20231213012653-1006025d24f8
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.17.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	google.golang.org/protobuf v1.36.8 // indirect
)

replace github.com/prometheus/common => github.com/prometheus/common v0.44.0

replace golang.org/x/sys => golang.org/x/sys v0.3.0
