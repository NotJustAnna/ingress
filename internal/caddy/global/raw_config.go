package global

import (
	"encoding/json"
	"fmt"

	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/caddyserver/ingress/pkg/converter"
	"github.com/caddyserver/ingress/pkg/store"
)

type RawConfigPlugin struct{}

func init() {
	converter.RegisterPlugin(RawConfigPlugin{})
}

func (p RawConfigPlugin) IngressPlugin() converter.PluginInfo {
	return converter.PluginInfo{
		Name: "raw_config",
		// Negative priority so it runs after the ingress plugin, which
		// overwrites (not appends to) the server's route list.
		Priority: -5,
		New:      func() converter.Plugin { return new(RawConfigPlugin) },
	}
}

// GlobalHandler appends the extraRoutes configmap option to the server routes
// and installs the errorRoutes configmap option as the server error handler.
func (p RawConfigPlugin) GlobalHandler(config *converter.Config, store *store.Store) error {
	cfgMap := store.ConfigMap
	httpServer := config.GetHTTPServer()

	if cfgMap.ExtraRoutes != "" {
		var routes caddyhttp.RouteList
		if err := json.Unmarshal([]byte(cfgMap.ExtraRoutes), &routes); err != nil {
			return fmt.Errorf("parsing extraRoutes: %w", err)
		}
		httpServer.Routes = append(httpServer.Routes, routes...)
	}

	if cfgMap.ErrorRoutes != "" {
		var routes caddyhttp.RouteList
		if err := json.Unmarshal([]byte(cfgMap.ErrorRoutes), &routes); err != nil {
			return fmt.Errorf("parsing errorRoutes: %w", err)
		}
		httpServer.Errors = &caddyhttp.HTTPErrorConfig{Routes: routes}
	}

	return nil
}

// Interface guards
var (
	_ = converter.GlobalMiddleware(RawConfigPlugin{})
)
