package validator

import (
	"encoding/json"
	"testing"

	"github.com/skip-mev/connect/v2/x/marketmap/types"
	"github.com/skip-mev/connect/v2/x/marketmap/types/tickermetadata"
	"github.com/stretchr/testify/require"
)

func TestGetCMCIDFromMetadata(t *testing.T) {
	testCases := []struct {
		name   string
		md     tickermetadata.CoreMetadata
		expErr error
		expID  int
	}{
		{
			name:   "empty",
			md:     tickermetadata.CoreMetadata{},
			expErr: ErrCMCIDNotFound,
		},
		{
			name: "random venue",
			md: tickermetadata.CoreMetadata{
				AggregateIDs: []tickermetadata.AggregatorID{{Venue: "FOO", ID: "325"}},
			},
			expErr: ErrCMCIDNotFound,
		},
		{
			name: "happy path",
			md: tickermetadata.CoreMetadata{
				AggregateIDs: []tickermetadata.AggregatorID{{Venue: "coinmarketcap", ID: "325"}},
			},
			expID: 325,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			id, err := getCMCIDFromMetadata(tc.md)
			if tc.expErr != nil {
				require.EqualError(t, err, tc.expErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expID, id)
			}
		})
	}
}

func TestGetCMCIDMapping(t *testing.T) {
	md1 := tickermetadata.CoreMetadata{AggregateIDs: []tickermetadata.AggregatorID{{Venue: "coinmarketcap", ID: "325"}}}
	bz, err := json.Marshal(md1)
	require.NoError(t, err)

	md2 := tickermetadata.CoreMetadata{AggregateIDs: []tickermetadata.AggregatorID{{Venue: "coinmarketcap", ID: "333"}}}
	bz2, err := json.Marshal(md2)
	require.NoError(t, err)

	mm := types.MarketMap{Markets: map[string]types.Market{
		"FOO/BAR": {
			Ticker: types.Ticker{Metadata_JSON: string(bz)},
		},
		"BAR/FOO": {
			Ticker: types.Ticker{Metadata_JSON: string(bz2)},
		},
	}}

	mapping, err := getCMCIDMapping(mm)
	require.NoError(t, err)

	require.Equal(t, mapping["FOO/BAR"], int64(325))
	require.Equal(t, mapping["BAR/FOO"], int64(333))
	require.NotContains(t, mapping, "ETH/USD")
}
