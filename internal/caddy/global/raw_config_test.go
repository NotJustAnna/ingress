package global

import (
	"testing"

	"github.com/caddyserver/ingress/pkg/converter"
	"github.com/caddyserver/ingress/pkg/store"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRawConfigRoutes(t *testing.T) {
	testCases := []struct {
		desc           string
		options        store.ConfigMapOptions
		expectedError  string
		expectedRoutes int
		expectedErrors int
	}{
		{
			desc:    "No options leaves server untouched",
			options: store.ConfigMapOptions{},
		},
		{
			desc: "Extra routes appended to server routes",
			options: store.ConfigMapOptions{
				ExtraRoutes: `[{"handle": [{"handler": "static_response", "status_code": 404, "body": "nothing here"}], "terminal": true}]`,
			},
			expectedRoutes: 1,
		},
		{
			desc: "Error routes installed on server errors",
			options: store.ConfigMapOptions{
				ErrorRoutes: `[{"handle": [{"handler": "static_response", "body": "site is down"}]}]`,
			},
			expectedErrors: 1,
		},
		{
			desc: "Invalid extra routes JSON",
			options: store.ConfigMapOptions{
				ExtraRoutes: `{"not": "an array"}`,
			},
			expectedError: "parsing extraRoutes: json: cannot unmarshal object into Go value of type caddyhttp.RouteList",
		},
		{
			desc: "Invalid error routes JSON",
			options: store.ConfigMapOptions{
				ErrorRoutes: `{"not": "an array"}`,
			},
			expectedError: "parsing errorRoutes: json: cannot unmarshal object into Go value of type caddyhttp.RouteList",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			p := RawConfigPlugin{}
			c := converter.NewConfig()
			s := store.NewStore(store.Options{}, "", &store.PodInfo{})
			*s.ConfigMap = tC.options

			err := p.GlobalHandler(c, s)

			if tC.expectedError != "" {
				require.EqualError(t, err, tC.expectedError)
				return
			}
			require.NoError(t, err)

			server := c.GetHTTPServer()
			assert.Len(t, server.Routes, tC.expectedRoutes)
			if tC.expectedErrors == 0 {
				assert.Nil(t, server.Errors)
			} else {
				require.NotNil(t, server.Errors)
				assert.Len(t, server.Errors.Routes, tC.expectedErrors)
			}
		})
	}
}

// The ingress plugin overwrites the server route list, so raw_config must be
// ordered after it for extraRoutes to survive as a trailing catch-all.
func TestRawConfigRunsAfterIngressPlugin(t *testing.T) {
	c := converter.NewConfig()
	s := store.NewStore(store.Options{}, "", &store.PodInfo{})
	s.ConfigMap.ExtraRoutes = `[{"handle": [{"handler": "static_response", "status_code": 404}], "terminal": true}]`
	s.AddIngress(&networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "example"},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{{
				Host: "example.com",
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{{
							Path:     "/",
							PathType: ptrTo(networkingv1.PathTypePrefix),
							Backend: networkingv1.IngressBackend{
								Service: &networkingv1.IngressServiceBackend{
									Name: "example",
									Port: networkingv1.ServiceBackendPort{Number: 8080},
								},
							},
						}},
					},
				},
			}},
		},
	})

	for _, plugin := range converter.Plugins(s.Options.PluginsOrder) {
		if m, ok := plugin.(converter.GlobalMiddleware); ok {
			require.NoError(t, m.GlobalHandler(c, s))
		}
	}

	routes := c.GetHTTPServer().Routes
	require.Len(t, routes, 2, "expected the ingress route followed by the extra route")
	assert.True(t, routes[1].Terminal, "extra catch-all route must come last")
}

func ptrTo[T any](v T) *T { return &v }
