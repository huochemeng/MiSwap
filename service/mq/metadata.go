package mq

import (
	"MiSwap/dto"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"MiSwap/base/stores/xkv"
	"github.com/pkg/errors"
)

const CacheRefreshSingleItemMetadataKey = "cache:%s:%s:item:refresh:metadata"

func GetRefreshSingleItemMetadataKey(project, chain string) string {
	return fmt.Sprintf(CacheRefreshSingleItemMetadataKey, strings.ToLower(project), strings.ToLower(chain))
}

const CacheRefreshPreventReentrancyKeyPrefix = "cache:es:item:refresh:prevent:reentrancy:%d:%s:%s"
const PreventReentrancyPeriod = 10 //second

func AddSingleItemToRefreshMetadataQueue(kvStore *xkv.Store, project, chainName string, chainID int64, collectionAddr, tokenID string) error {
	isRefreshed, err := kvStore.Get(fmt.Sprintf(CacheRefreshPreventReentrancyKeyPrefix, chainID, collectionAddr, tokenID))
	if err != nil {
		return errors.Wrap(err, "failed on check reentrancy status")
	}

	if isRefreshed != "" {
		slog.InfoContext(context.Background(), "refresh within 10s", "collection_address", collectionAddr, "token_id", tokenID)
		return nil
	}

	item := dto.RefreshItem{
		ChainID:        chainID,
		CollectionAddr: collectionAddr,
		TokenID:        tokenID,
	}

	rawInfo, err := json.Marshal(&item)
	if err != nil {
		return errors.Wrap(err, "failed on marshal item info")
	}

	_, err = kvStore.Sadd(GetRefreshSingleItemMetadataKey(project, chainName), string(rawInfo))
	if err != nil {
		return errors.Wrap(err, "failed to push item to refresh metadata queue")
	}

	_ = kvStore.Setex(fmt.Sprintf(CacheRefreshPreventReentrancyKeyPrefix, chainID, collectionAddr, tokenID), "true", PreventReentrancyPeriod)

	return nil
}
