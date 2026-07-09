package ingress

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/caddyserver/ingress/pkg/converter"
	"github.com/stretchr/testify/assert"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRawHandlersConvertToCaddyConfig(t *testing.T) {
	rp := RawHandlersPlugin{}

	tests := []struct {
		name               string
		expectedConfigPath string
		expectedError      string
		annotations        map[string]string
	}{
		{
			name:               "No annotation leaves route untouched",
			expectedConfigPath: "test_data/rawhandlers_empty.json",
			annotations:        map[string]string{},
		},
		{
			name:               "Raw handlers appended to route",
			expectedConfigPath: "test_data/rawhandlers_static_response.json",
			annotations: map[string]string{
				"caddy.ingress.kubernetes.io/raw-handlers": `[
					{"handler": "headers", "response": {"set": {"X-Powered-By": ["caddy"]}}},
					{"handler": "static_response", "status_code": 200, "body": "hello"}
				]`,
			},
		},
		{
			name: "Invalid JSON returns an error",
			annotations: map[string]string{
				"caddy.ingress.kubernetes.io/raw-handlers": `{"handler": "not-an-array"}`,
			},
			expectedError: "parsing raw-handlers annotation on ingress default/example: json: cannot unmarshal object into Go value of type []json.RawMessage",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := converter.IngressMiddlewareInput{
				Ingress: &networkingv1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:   "default",
						Name:        "example",
						Annotations: test.annotations,
					},
				},
				Route: &caddyhttp.Route{},
			}

			route, err := rp.IngressHandler(input)

			if test.expectedError != "" {
				if assert.Error(t, err, "expected an error while generating the ingress route") {
					assert.EqualError(t, err, test.expectedError)
				}
				return
			}
			assert.NoError(t, err, "failed to generate ingress route")

			expectedCfg, err := os.ReadFile(test.expectedConfigPath)
			assert.NoError(t, err, "failed to find config file for comparison")

			cfgJSON, err := json.Marshal(&route)
			assert.NoError(t, err, "failed to marshal route to JSON")

			assert.JSONEq(t, string(cfgJSON), string(expectedCfg))
		})
	}
}
