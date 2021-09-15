package input

import (
	"context"
	"fmt"
	"io"

	"github.com/stellar/go/ingest"
	"github.com/stellar/go/ingest/ledgerbackend"
	"github.com/stellar/go/network"
	"github.com/stellar/go/xdr"
	"github.com/stellar/stellar-etl/internal/utils"
)

// OperationTransformInput is a representation of the input for the TransformOperation function
type OperationTransformInputCaptive struct {
	Operation      xdr.Operation
	OperationIndex int32
	Transaction    ingest.LedgerTransaction
	LedgerSeqNum   int32
}

func panicIf(err error) {
	if err != nil {
		panic(fmt.Errorf("An error occurred, panicking: %s\n", err))
	}
}

// GetOperations returns a slice of operations for the ledgers in the provided range (inclusive on both ends)
func GetOperationsCaptive(start, end uint32, limit int64) ([]OperationTransformInput, error) {
	ctx := context.Background()
	captiveCoreToml, err := ledgerbackend.NewCaptiveCoreTomlFromFile(
		"docker/stellar-core.cfg",
		ledgerbackend.CaptiveCoreTomlParams{
			NetworkPassphrase:  network.PublicNetworkPassphrase,
			HistoryArchiveURLs: utils.ArchiveURLs,
			Strict:             true,
		},
	)

	backend, err := ledgerbackend.NewCaptive(
		ledgerbackend.CaptiveCoreConfig{
			BinaryPath:         "/usr/bin/stellar-core",
			Toml:               captiveCoreToml,
			NetworkPassphrase:  network.PublicNetworkPassphrase,
			HistoryArchiveURLs: utils.ArchiveURLs,
		},
	)
	if err != nil {
		return []OperationTransformInput{}, err
	}

	opSlice := []OperationTransformInput{}
	err = backend.PrepareRange(ctx, ledgerbackend.BoundedRange(start, end))
	panicIf(err)
	for seq := start; seq <= end; seq++ {
		txReader, err := ingest.NewLedgerTransactionReader(ctx, backend, publicPassword, seq)
		if err != nil {
			return []OperationTransformInput{}, err
		}

		for int64(len(opSlice)) < limit || limit < 0 {
			tx, err := txReader.Read()
			if err == io.EOF {
				break
			}

			for index, op := range tx.Envelope.Operations() {
				opSlice = append(opSlice, OperationTransformInput{
					Operation:      op,
					OperationIndex: int32(index),
					Transaction:    tx,
					LedgerSeqNum:   int32(seq),
				})

				if int64(len(opSlice)) >= limit && limit >= 0 {
					break
				}
			}
		}

		txReader.Close()
		if int64(len(opSlice)) >= limit && limit >= 0 {
			break
		}
	}

	return opSlice, nil
}
