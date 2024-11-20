package simulate

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAgentConfig_Validate(t *testing.T) {
	testCases := []struct {
		name   string
		cfg    SigningAgentConfig
		expErr bool
	}{
		{
			name:   "empty config invalid",
			cfg:    SigningAgentConfig{},
			expErr: true,
		},
		{
			name: "valid config",
			cfg: SigningAgentConfig{
				Address: "cosmos1abcd",
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			require.Equal(t, tc.expErr, err != nil, "test %q failed: %s", tc.name, err)
		})
	}
}
