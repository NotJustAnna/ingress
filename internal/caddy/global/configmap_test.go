package global

import (
	"encoding/json"
	"testing"

	"github.com/caddyserver/ingress/pkg/converter"
	"github.com/caddyserver/ingress/pkg/store"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigMapDNSProvider(t *testing.T) {
	testCases := []struct {
		desc             string
		options          store.ConfigMapOptions
		expectedError    string
		expectedIssuers  []string
		expectedPolicies int
	}{
		{
			desc:    "No automation without acmeCA or email",
			options: store.ConfigMapOptions{},
		},
		{
			desc: "DNS provider set on default issuer",
			options: store.ConfigMapOptions{
				Email:       "test@example.com",
				DNSProvider: `{"name":"cloudflare","api_token":"{env.CF_API_TOKEN}"}`,
			},
			expectedIssuers: []string{
				`{"challenges":{"dns":{"provider":{"name":"cloudflare","api_token":"{env.CF_API_TOKEN}"}}},"email":"test@example.com","module":"acme"}`,
			},
			expectedPolicies: 1,
		},
		{
			desc: "Invalid DNS provider JSON",
			options: store.ConfigMapOptions{
				Email:       "test@example.com",
				DNSProvider: `{name: cloudflare}`,
			},
			expectedError: "dnsProvider is not valid JSON: {name: cloudflare}",
		},
		{
			desc: "Extra automation policies appended after default policy",
			options: store.ConfigMapOptions{
				Email: "test@example.com",
				ExtraTLSAutomationPolicies: `[{
					"subjects": ["*.example.com"],
					"issuers": [{"module": "acme", "challenges": {"dns": {"provider": {"name": "route53"}}}}]
				}]`,
			},
			expectedIssuers:  []string{`{"email":"test@example.com","module":"acme"}`},
			expectedPolicies: 2,
		},
		{
			desc: "Extra automation policies without default issuer",
			options: store.ConfigMapOptions{
				ExtraTLSAutomationPolicies: `[{"subjects": ["*.example.com"]}]`,
			},
			expectedPolicies: 1,
		},
		{
			desc: "Invalid extra automation policies JSON",
			options: store.ConfigMapOptions{
				ExtraTLSAutomationPolicies: `{"not": "an array"}`,
			},
			expectedError: "parsing extraTLSAutomationPolicies: json: cannot unmarshal object into Go value of type []*caddytls.AutomationPolicy",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			p := ConfigMapPlugin{}
			c := converter.NewConfig()
			s := store.NewStore(store.Options{}, "", &store.PodInfo{})
			*s.ConfigMap = tC.options

			err := p.GlobalHandler(c, s)

			if tC.expectedError != "" {
				require.EqualError(t, err, tC.expectedError)
				return
			}
			require.NoError(t, err)

			automation := c.GetTLSApp().Automation
			if tC.expectedPolicies == 0 {
				assert.Nil(t, automation)
				return
			}
			require.NotNil(t, automation)
			require.Len(t, automation.Policies, tC.expectedPolicies)

			if len(tC.expectedIssuers) > 0 {
				require.Len(t, automation.Policies[0].IssuersRaw, len(tC.expectedIssuers))
				for i, expected := range tC.expectedIssuers {
					issuer, err := json.Marshal(automation.Policies[0].IssuersRaw[i])
					require.NoError(t, err)
					assert.JSONEq(t, expected, string(issuer))
				}
			}
		})
	}
}
