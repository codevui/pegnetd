package srv

import (
	"fmt"
	"strings"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/pegnet/pegnetd/fat/fat2"
	"github.com/pegnet/pegnetd/node"
)

func signAndSend(tx *fat2.Transaction, cl *factom.Client, payment string) (err error, commit *factom.Bytes32, reveal *factom.Bytes32) {
	// Get out private key
	fmt.Printf("sign and send 1\n")
	priv, err := tx.Input.Address.GetFsAddress(cl)
	if err != nil {
		return fmt.Errorf("unable to get private key: %s\n", err.Error()), nil, nil
	}
	fmt.Printf("sign and send 2\n")

	var txBatch fat2.TransactionBatch
	txBatch.Version = 1
	txBatch.Transactions = []fat2.Transaction{*tx}
	txBatch.ChainID = &node.TransactionChain

	// Sign the tx and make an entry
	err = txBatch.MarshalEntry()
	if err != nil {
		return fmt.Errorf("failed to marshal tx: %s", err.Error()), nil, nil
	}

	txBatch.Sign(priv)
	fmt.Printf("sign and send 3\n")

	if err := txBatch.Validate(); err != nil {
		return fmt.Errorf("invalid tx: %s", err.Error()), nil, nil
	}
	fmt.Printf("sign and send 4\n")
	ec, err := factom.NewECAddress(payment)
	if err != nil {
		return fmt.Errorf("failed to parse input: %s\n", err.Error()), nil, nil
	}
	fmt.Printf("sign and send 5\n")
	bal, err := ec.GetBalance(cl)
	if err != nil {
		return fmt.Errorf("failed to get ec balance: %s\n", err.Error()), nil, nil
	}
	fmt.Printf("sign and send 6\n")
	if cost, err := txBatch.Cost(false); err != nil || uint64(cost) > bal {
		return fmt.Errorf("not enough ec balance for the transaction"), nil, nil
	}
	fmt.Printf("sign and send 7\n")
	es, err := ec.GetEsAddress(cl)
	if err != nil {
		return fmt.Errorf("failed to parse input: %s\n", err.Error()), nil, nil
	}
	fmt.Printf("sign and send 8\n")
	txid, err := txBatch.ComposeCreate(cl, es, false)
	if err != nil {
		return fmt.Errorf("failed to submit entry: %s\n", err.Error()), nil, nil
	}
	fmt.Printf("sign and send 9\n")
	return nil, txid, txBatch.Hash
}

func setTransferOutput(tx *fat2.Transaction, dest, amt string) error {
	var err error
	amount, err := FactoidToFactoshi(amt)
	if err != nil {
		return fmt.Errorf("invalid amount specified: %s\n", err.Error())
	}

	tx.Transfers = make([]fat2.AddressAmountTuple, 1)
	tx.Transfers[0].Amount = uint64(amount)
	if tx.Transfers[0].Address, err = factom.NewFAAddress(dest); err != nil {
		return fmt.Errorf("failed to parse input: %s\n", err.Error())
	}

	return nil
}

func setTransactionInput(tx *fat2.Transaction, source, asset, amt string) error {
	var err error
	if tx.Input.Type, err = ticker(asset); err != nil {
		return err
	}

	amount, err := FactoidToFactoshi(amt)
	if err != nil {
		return fmt.Errorf("invalid amount specified: %s\n", err.Error())
	}
	tx.Input.Amount = uint64(amount)

	// Set the input
	if tx.Input.Address, err = factom.NewFAAddress(source); err != nil {
		return fmt.Errorf("failed to parse input: %s\n", err.Error())
	}

	//pBals, err := srv.getPegnetBalances(source)
	//if err != nil {
	//	return fmt.Errorf("failed to get asset balance: %s", err.Error())
	//}
	//
	//if pBals[tx.Input.Type] < tx.Input.Amount {
	//	return fmt.Errorf("not enough %s to cover the transaction", tx.Input.Type.String())
	//}

	return nil
}

func ticker(asset string) (fat2.PTicker, error) {
	// No asset starts with a 'p', so we can do the quick check
	// if the start is a p for if it is already in 'p' form.
	// TODO: Make a more robust Asset -> pAsset converter
	if strings.ToUpper(asset) != "PEG" && asset[0] != 'p' {
		asset = "p" + strings.ToUpper(asset)
	}
	aType := fat2.StringToTicker(asset)
	if aType == fat2.PTickerInvalid {
		return fat2.PTickerInvalid, fmt.Errorf("invalid ticker type\n")
	}
	return aType, nil
}
