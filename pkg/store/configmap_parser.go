package store

import (
	"fmt"
	"reflect"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/mitchellh/mapstructure"
	apiv1 "k8s.io/api/core/v1"
)

// ConfigMapOptions represents global options set through a configmap
type ConfigMapOptions struct {
	Debug                 bool           `json:"debug,omitempty"`
	AcmeCA                string         `json:"acmeCA,omitempty"`
	AcmeEABKeyID          string         `json:"acmeEABKeyId,omitempty"`
	AcmeEABMacKey         string         `json:"acmeEABMacKey,omitempty"`
	Email                 string         `json:"email,omitempty"`
	ExperimentalSmartSort bool           `json:"experimentalSmartSort,omitempty"`
	ProxyProtocol         bool           `json:"proxyProtocol,omitempty"`
	Metrics               bool           `json:"metrics,omitempty"`
	OnDemandTLS           bool           `json:"onDemandTLS,omitempty"`
	OnDemandAsk           string         `json:"onDemandAsk,omitempty"`
	OCSPCheckInterval     caddy.Duration `json:"ocspCheckInterval,omitempty"`

	// ExtraRoutes is a raw JSON array of http routes appended after all
	// ingress-generated routes, e.g. a catch-all route serving a static
	// response for otherwise unmatched hosts. Referenced handler modules
	// must be compiled into the controller binary.
	ExtraRoutes string `json:"extraRoutes,omitempty"`
	// ErrorRoutes is a raw JSON array of http routes installed as the
	// server's error handling routes (the JSON equivalent of the Caddyfile
	// handle_errors directive), e.g. to serve a friendly page when a
	// backend is unreachable instead of an empty 502.
	ErrorRoutes string `json:"errorRoutes,omitempty"`
}

func stringToCaddyDurationHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(caddy.Duration(time.Second)) {
			return data, nil
		}
		return caddy.ParseDuration(data.(string))
	}
}

func ParseConfigMap(cm *apiv1.ConfigMap) (*ConfigMapOptions, error) {
	// parse configmap
	cfgMap := ConfigMapOptions{}
	config := &mapstructure.DecoderConfig{
		Metadata:         nil,
		WeaklyTypedInput: true,
		Result:           &cfgMap,
		TagName:          "json",
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			stringToCaddyDurationHookFunc(),
		),
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return nil, fmt.Errorf("unexpected error creating decoder: %w", err)
	}
	err = decoder.Decode(cm.Data)
	if err != nil {
		return nil, fmt.Errorf("unexpected error parsing configmap: %w", err)
	}

	return &cfgMap, nil
}
