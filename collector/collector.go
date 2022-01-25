package collector

import (
	"context"
	"time"

	"radius-exporter/config"

	"github.com/prometheus/client_golang/prometheus"

	log "github.com/sirupsen/logrus"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

const (
	namespace = "radius"
)

type Collector struct {
	Target       *string
	Module       *config.Module
	duration     prometheus.Gauge
	responseCode prometheus.Gauge
	success      prometheus.Gauge
}

func NewCollector(target *string, module *config.Module) Collector {
	return Collector{
		Target: target,
		Module: module,
		duration: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "scrape_duration_seconds",
			Help:      "RADIUS response time in seconds.",
		}),
		responseCode: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "response_code",
			Help:      "The RADIUS response code. Common values are Access-Accept(2) and Access-Reject(3).",
		}),
		success: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "success",
			Help:      "1 if the radius probe was successful.",
		}),
	}
}

// Describe implements prometheus.Collector
func (c Collector) Describe(ch chan<- *prometheus.Desc) {
	c.duration.Describe(ch)
	c.responseCode.Describe(ch)
	c.success.Describe(ch)
}

// Collect implements prometheus.Collector
func (c Collector) Collect(ch chan<- prometheus.Metric) {
	err := c.probe()
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"target": *c.Target,
		}).Error("Probe failure. Error sending radius request.")
		c.success.Set(0)
	} else {
		log.WithFields(log.Fields{"target": *c.Target}).Debug("Probe success.")
		c.success.Set(1)
	}

	c.duration.Collect(ch)
	c.responseCode.Collect(ch)
	c.success.Collect(ch)
}

func (c Collector) probe() error {
	client := radius.Client{Retry: c.Module.Retry, MaxPacketErrors: c.Module.MaxPacketErrors}

	packet := radius.New(radius.CodeAccessRequest, c.Module.Secret)
	err := rfc2865.UserName_SetString(packet, c.Module.Username)
	if err != nil {
		return err
	}

	err = rfc2865.UserPassword_SetString(packet, c.Module.Password)
	if err != nil {
		return err
	}

	if c.Module.NasID != "" {
		err = rfc2865.NASIdentifier_SetString(packet, c.Module.NasID)
		if err != nil {
			return err
		}
	}
	if c.Module.NasIP != nil {
		err = rfc2865.NASIPAddress_Add(packet, c.Module.NasIP)
		if err != nil {
			return err
		}
	}

	begin := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), c.Module.Timeout)
	defer cancel()
	response, err := client.Exchange(ctx, packet, *c.Target)

	if err != nil {
		return err
	}
	c.responseCode.Set((float64)(response.Code))
	c.duration.Set(time.Since(begin).Seconds())
	return nil
}
