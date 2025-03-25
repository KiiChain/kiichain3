package client

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/rs/zerolog"
)

// AccountInfo handle the account sequence
type AccountInfo struct {
	AccountNumber       uint64
	AccountSequence     uint64
	ShouldResetSequence bool
}

// NewAccountInfo creates a new instance of AccountInfo
func NewAccountInfo() *AccountInfo {
	return &AccountInfo{}
}

// ObtainAccountInfo ensures the account defined by ctx.GetFromAddress() exists.
// We keep a local copy of account sequence number and manually increment it.
// If the local sequence number is 0, we will initialize it with the latest value getting from the chain.
func (accountInfo *AccountInfo) ObtainAccountInfo(ctx client.Context, txf tx.Factory, logger zerolog.Logger) (tx.Factory, error) {
	// reset the account sequence
	if accountInfo.AccountSequence == 0 || accountInfo.ShouldResetSequence {
		err := accountInfo.ResetAccountSequence(ctx, txf, logger)
		if err != nil {
			return txf, err
		}
		accountInfo.ShouldResetSequence = false
	}

	// set the account sequence on the built transaction
	txf = txf.WithAccountNumber(accountInfo.AccountNumber)
	txf = txf.WithSequence(accountInfo.AccountSequence)
	txf = txf.WithGas(0)
	return txf, nil
}

// ResetAccountSequence will reset account sequence number to the latest sequence number in the chain
func (accountInfo *AccountInfo) ResetAccountSequence(ctx client.Context, txf tx.Factory, logger zerolog.Logger) error {
	// get the "From" of the transaction
	fromAddr := ctx.GetFromAddress()

	// validate the from account exists
	err := txf.AccountRetriever().EnsureExists(ctx, fromAddr)
	if err != nil {
		return err
	}

	// get account number and current sequence
	accountNum, sequence, err := txf.AccountRetriever().GetAccountNumberSequence(ctx, fromAddr)
	if err != nil {
		return err
	}

	// set account number and sequence on the struct
	logger.Info().Msg(fmt.Sprintf("Reset account number to %d and sequence number to %d", accountNum, sequence))
	accountInfo.AccountNumber = accountNum
	accountInfo.AccountSequence = sequence
	return nil
}
