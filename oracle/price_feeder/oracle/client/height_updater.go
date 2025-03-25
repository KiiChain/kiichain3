package client

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	tmrpcclient "github.com/tendermint/tendermint/rpc/client"
	tmtypes "github.com/tendermint/tendermint/types"
)

var (
	started                  = false
	queryEventNewBlockHeader = tmtypes.QueryForEvent(tmtypes.EventNewBlockHeaderValue) // event to be queried
	queryInterval            = 20 * time.Millisecond                                   // time between query the latest new block event
)

// HeightUpdater is used to provide the updates of the latest chain
// It starts a goroutine to subscribe to new block event and send the latest block height to the channel
type HeightUpdater struct {
	Logger        zerolog.Logger
	LastHeight    int64 // store the last processed block height
	ChBlockHeight chan int64
}

// Start starts the rpc client and subscribes to EventNewBlockHeader.
func (heightUpdater HeightUpdater) Start(
	ctx context.Context,
	rpcClient tmrpcclient.Client,
	logger zerolog.Logger,
) error {
	if !started {
		// start rpc connection
		err := rpcClient.Start(ctx)
		if err != nil {
			return err
		}

		// track the new block events generated and update the chain height
		go heightUpdater.subscribe(ctx, rpcClient, logger)
		started = true
	}
	return nil
}

// subscribe listens to new blocks being made
// and updates the chain height.
func (heightUpdater HeightUpdater) subscribe(
	ctx context.Context,
	eventsClient tmrpcclient.EventsClient,
	logger zerolog.Logger,
) {
	for {
		// wait until a EventNewBlockHeader event
		eventData, err := tmrpcclient.WaitForOneEvent(ctx, eventsClient, queryEventNewBlockHeader.String())
		if err != nil {
			logger.Debug().Err(err).Msg("Failed to query EventNewBlockHeader")
		}

		// check if the event received is type EventDataNewBlockHeader
		eventDataNewBlockHeader, ok := eventData.(tmtypes.EventDataNewBlockHeader)
		if !ok {
			logger.Err(err).Msg("Failed to parse event from eventDataNewBlockHeader")
			continue
		}

		// extract the block height from the event
		eventHeight := eventDataNewBlockHeader.Header.Height
		if eventHeight > heightUpdater.LastHeight {
			logger.Info().Msg(fmt.Sprintf("Received new Chain Height: %d", eventHeight))
			heightUpdater.LastHeight = eventHeight // update the height with the latest

			if len(heightUpdater.ChBlockHeight) < 1 {
				heightUpdater.ChBlockHeight <- eventHeight // update the height on the channel
			} else {
				// skip this block height since price feeder is still sending previous transaction
				logger.Info().Msg(fmt.Sprintf("Skipped Block Height: %d due to in progress tx", eventHeight))
			}
		}

		time.Sleep(queryInterval)
	}
}
