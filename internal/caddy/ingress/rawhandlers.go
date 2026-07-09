package ingress

import (
	"encoding/json"
	"fmt"

	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/caddyserver/ingress/pkg/converter"
)

type RawHandlersPlugin struct{}

func (p RawHandlersPlugin) IngressPlugin() converter.PluginInfo {
	return converter.PluginInfo{
		Name: "ingress.rawhandlers",
		// After redirect/rewrite (10), before reverseproxy (-10) so raw
		// handlers can short-circuit or modify requests to the backend.
		Priority: 5,
		New:      func() converter.Plugin { return new(RawHandlersPlugin) },
	}
}

// IngressHandler appends handlers from the raw-handlers annotation, a JSON
// array of http.handlers module objects. Referenced handler modules must be
// compiled into the controller binary.
func (p RawHandlersPlugin) IngressHandler(input converter.IngressMiddlewareInput) (*caddyhttp.Route, error) {
	raw := getAnnotation(input.Ingress, rawHandlersAnnotation)
	if raw == "" {
		return input.Route, nil
	}

	var handlers []json.RawMessage
	if err := json.Unmarshal([]byte(raw), &handlers); err != nil {
		return nil, fmt.Errorf("parsing %s annotation on ingress %s/%s: %w",
			rawHandlersAnnotation, input.Ingress.Namespace, input.Ingress.Name, err)
	}

	input.Route.HandlersRaw = append(input.Route.HandlersRaw, handlers...)
	return input.Route, nil
}

func init() {
	converter.RegisterPlugin(RawHandlersPlugin{})
}

// Interface guards
var (
	_ = converter.IngressMiddleware(RawHandlersPlugin{})
)
