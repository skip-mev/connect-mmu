package util

import (
	"github.com/cosmos/gogoproto/proto"
	"golang.org/x/exp/maps"

	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
)

func GenerateUpdate(mm mmtypes.MarketMap, authority string) proto.Message {
	msg := mmtypes.MsgUpdateMarkets{
		Authority:     authority,
		UpdateMarkets: maps.Values(mm.Markets),
	}

	return &msg
}

func GenerateUpsert(mm mmtypes.MarketMap, authority string) proto.Message {
	msg := mmtypes.MsgUpsertMarkets{
		Authority: authority,
		Markets:   maps.Values(mm.Markets),
	}

	return &msg
}

func GenerateCreate(mm mmtypes.MarketMap, authority string) proto.Message {
	msg := mmtypes.MsgCreateMarkets{
		Authority:     authority,
		CreateMarkets: maps.Values(mm.Markets),
	}

	return &msg
}
