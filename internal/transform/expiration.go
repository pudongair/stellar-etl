package transform

import (
	"fmt"

	"github.com/stellar/go/ingest"
	"github.com/stellar/go/strkey"
	"github.com/stellar/go/xdr"
	"github.com/stellar/stellar-etl/internal/utils"
)

// TransformConfigSetting converts an config setting ledger change entry into a form suitable for BigQuery
func TransformExpiration(ledgerChange ingest.Change, header xdr.LedgerHeaderHistoryEntry) (ExpirationOutput, error) {
	ledgerEntry, changeType, outputDeleted, err := utils.ExtractEntryFromChange(ledgerChange)
	if err != nil {
		return ExpirationOutput{}, err
	}

	expiration, ok := ledgerEntry.Data.GetExpiration()
	if !ok {
		return ExpirationOutput{}, fmt.Errorf("Could not extract expiration from ledger entry; actual type is %s", ledgerEntry.Data.Type)
	}

	// LedgerEntryChange must contain a expiration change to be parsed, otherwise skip
	if ledgerEntry.Data.Type != xdr.LedgerEntryTypeExpiration {
		return ExpirationOutput{}, nil
	}

	keyHashByte, _ := expiration.KeyHash.MarshalBinary()
	keyHash, _ := strkey.Encode(strkey.VersionByteContract, keyHashByte)
	expirationLedgerSeq := expiration.ExpirationLedgerSeq

	closedAt, err := utils.TimePointToUTCTimeStamp(header.Header.ScpValue.CloseTime)
	if err != nil {
		return ExpirationOutput{}, err
	}

	ledgerSequence := header.Header.LedgerSeq

	transformedPool := ExpirationOutput{
		KeyHash:             keyHash,
		ExpirationLedgerSeq: uint32(expirationLedgerSeq),
		LastModifiedLedger:  uint32(ledgerEntry.LastModifiedLedgerSeq),
		LedgerEntryChange:   uint32(changeType),
		Deleted:             outputDeleted,
		ClosedAt:            closedAt,
		LedgerSequence:      uint32(ledgerSequence),
	}

	return transformedPool, nil
}
