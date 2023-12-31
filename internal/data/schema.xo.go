// Package data contains generated code for schema 'public'.
package data

// Code generated by xo. DO NOT EDIT.

import (
	"database/sql"
	"database/sql/driver"
	"encoding/csv"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/rarimo/xo/types/xo"
)

// StringSlice is a slice of strings.
type StringSlice []string

// quoteEscapeRegex is the regex to match escaped characters in a string.
var quoteEscapeRegex = regexp.MustCompile(`([^\\]([\\]{2})*)\\"`)

// Scan satisfies the sql.Scanner interface for StringSlice.
func (ss *StringSlice) Scan(src interface{}) error {
	buf, ok := src.([]byte)
	if !ok {
		return errors.New("invalid StringSlice")
	}

	// change quote escapes for csv parser
	str := quoteEscapeRegex.ReplaceAllString(string(buf), `$1""`)
	str = strings.Replace(str, `\\`, `\`, -1)

	// remove braces
	str = str[1 : len(str)-1]

	// bail if only one
	if len(str) == 0 {
		*ss = StringSlice([]string{})
		return nil
	}

	// parse with csv reader
	cr := csv.NewReader(strings.NewReader(str))
	slice, err := cr.Read()
	if err != nil {
		fmt.Printf("exiting!: %v\n", err)
		return err
	}

	*ss = StringSlice(slice)

	return nil
}

// Value satisfies the driver.Valuer interface for StringSlice.
func (ss StringSlice) Value() (driver.Value, error) {
	v := make([]string, len(ss))
	for i, s := range ss {
		v[i] = `"` + strings.Replace(strings.Replace(s, `\`, `\\\`, -1), `"`, `\"`, -1) + `"`
	}
	return "{" + strings.Join(v, ",") + "}", nil
} // Approval represents a row from 'public.approvals'.
type Approval struct {
	ID                int64     `db:"id" json:"id" structs:"-"`                                                  // id
	TransferIndex     []byte    `db:"transfer_index" json:"transfer_index" structs:"transfer_index"`             // transfer_index
	RarimoTransaction []byte    `db:"rarimo_transaction" json:"rarimo_transaction" structs:"rarimo_transaction"` // rarimo_transaction
	CreatedAt         time.Time `db:"created_at" json:"created_at" structs:"created_at"`                         // created_at

}

// Collection represents a row from 'public.collections'.
type Collection struct {
	ID        int64     `db:"id" json:"id" structs:"-"`                          // id
	Index     []byte    `db:"index" json:"index" structs:"index"`                // index
	Metadata  xo.Jsonb  `db:"metadata" json:"metadata" structs:"metadata"`       // metadata
	CreatedAt time.Time `db:"created_at" json:"created_at" structs:"created_at"` // created_at
	UpdatedAt time.Time `db:"updated_at" json:"updated_at" structs:"updated_at"` // updated_at

}

// CollectionChainMapping represents a row from 'public.collection_chain_mappings'.
type CollectionChainMapping struct {
	Collection int64         `db:"collection" json:"collection" structs:"-"`          // collection
	Network    int           `db:"network" json:"network" structs:"-"`                // network
	Address    []byte        `db:"address" json:"address" structs:"address"`          // address
	TokenType  sql.NullInt64 `db:"token_type" json:"token_type" structs:"token_type"` // token_type
	Wrapped    sql.NullBool  `db:"wrapped" json:"wrapped" structs:"wrapped"`          // wrapped
	Decimals   sql.NullInt64 `db:"decimals" json:"decimals" structs:"decimals"`       // decimals
	CreatedAt  time.Time     `db:"created_at" json:"created_at" structs:"created_at"` // created_at
	UpdatedAt  time.Time     `db:"updated_at" json:"updated_at" structs:"updated_at"` // updated_at

}

// Confirmation represents a row from 'public.confirmations'.
type Confirmation struct {
	ID                int64     `db:"id" json:"id" structs:"-"`                                                  // id
	TransferIndex     []byte    `db:"transfer_index" json:"transfer_index" structs:"transfer_index"`             // transfer_index
	RarimoTransaction []byte    `db:"rarimo_transaction" json:"rarimo_transaction" structs:"rarimo_transaction"` // rarimo_transaction
	CreatedAt         time.Time `db:"created_at" json:"created_at" structs:"created_at"`                         // created_at

}

// GorpMigration represents a row from 'public.gorp_migrations'.
type GorpMigration struct {
	ID        string       `db:"id" json:"id" structs:"-"`                          // id
	AppliedAt sql.NullTime `db:"applied_at" json:"applied_at" structs:"applied_at"` // applied_at

}

// Item represents a row from 'public.items'.
type Item struct {
	ID         int64         `db:"id" json:"id" structs:"-"`                          // id
	Index      []byte        `db:"index" json:"index" structs:"index"`                // index
	Collection sql.NullInt64 `db:"collection" json:"collection" structs:"collection"` // collection
	Metadata   xo.Jsonb      `db:"metadata" json:"metadata" structs:"metadata"`       // metadata
	CreatedAt  time.Time     `db:"created_at" json:"created_at" structs:"created_at"` // created_at
	UpdatedAt  time.Time     `db:"updated_at" json:"updated_at" structs:"updated_at"` // updated_at

}

// ItemChainMapping represents a row from 'public.item_chain_mappings'.
type ItemChainMapping struct {
	Item      int64     `db:"item" json:"item" structs:"-"`                      // item
	Network   int       `db:"network" json:"network" structs:"-"`                // network
	Address   []byte    `db:"address" json:"address" structs:"address"`          // address
	TokenID   []byte    `db:"token_id" json:"token_id" structs:"token_id"`       // token_id
	CreatedAt time.Time `db:"created_at" json:"created_at" structs:"created_at"` // created_at
	UpdatedAt time.Time `db:"updated_at" json:"updated_at" structs:"updated_at"` // updated_at

}

// Rejection represents a row from 'public.rejections'.
type Rejection struct {
	ID                int64     `db:"id" json:"id" structs:"-"`                                                  // id
	TransferIndex     []byte    `db:"transfer_index" json:"transfer_index" structs:"transfer_index"`             // transfer_index
	RarimoTransaction []byte    `db:"rarimo_transaction" json:"rarimo_transaction" structs:"rarimo_transaction"` // rarimo_transaction
	CreatedAt         time.Time `db:"created_at" json:"created_at" structs:"created_at"`                         // created_at

}

// Transaction represents a row from 'public.transactions'.
type Transaction struct {
	Hash        []byte        `db:"hash" json:"hash" structs:"-"`                            // hash
	BlockHeight sql.NullInt64 `db:"block_height" json:"block_height" structs:"block_height"` // block_height
	Index       sql.NullInt64 `db:"index" json:"index" structs:"index"`                      // index
	RawTx       []byte        `db:"raw_tx" json:"raw_tx" structs:"raw_tx"`                   // raw_tx
	TxResult    xo.NullJsonb  `db:"tx_result" json:"tx_result" structs:"tx_result"`          // tx_result
	TxTimestamp time.Time     `db:"tx_timestamp" json:"tx_timestamp" structs:"tx_timestamp"` // tx_timestamp
	CreatedAt   time.Time     `db:"created_at" json:"created_at" structs:"created_at"`       // created_at

}

// Transfer represents a row from 'public.transfers'.
type Transfer struct {
	ID                int64          `db:"id" json:"id" structs:"-"`                                                     // id
	Index             []byte         `db:"index" json:"index" structs:"index"`                                           // index
	Status            int            `db:"status" json:"status" structs:"status"`                                        // status
	CreatedAt         time.Time      `db:"created_at" json:"created_at" structs:"created_at"`                            // created_at
	UpdatedAt         time.Time      `db:"updated_at" json:"updated_at" structs:"updated_at"`                            // updated_at
	Creator           sql.NullString `db:"creator" json:"creator" structs:"creator"`                                     // creator
	RarimoTx          []byte         `db:"rarimo_tx" json:"rarimo_tx" structs:"rarimo_tx"`                               // rarimo_tx
	RarimoTxTimestamp time.Time      `db:"rarimo_tx_timestamp" json:"rarimo_tx_timestamp" structs:"rarimo_tx_timestamp"` // rarimo_tx_timestamp
	Origin            string         `db:"origin" json:"origin" structs:"origin"`                                        // origin
	Tx                []byte         `db:"tx" json:"tx" structs:"tx"`                                                    // tx
	EventID           int64          `db:"event_id" json:"event_id" structs:"event_id"`                                  // event_id
	FromChain         string         `db:"from_chain" json:"from_chain" structs:"from_chain"`                            // from_chain
	ToChain           string         `db:"to_chain" json:"to_chain" structs:"to_chain"`                                  // to_chain
	Receiver          string         `db:"receiver" json:"receiver" structs:"receiver"`                                  // receiver
	Amount            Int256         `db:"amount" json:"amount" structs:"amount"`                                        // amount
	BundleData        []byte         `db:"bundle_data" json:"bundle_data" structs:"bundle_data"`                         // bundle_data
	BundleSalt        []byte         `db:"bundle_salt" json:"bundle_salt" structs:"bundle_salt"`                         // bundle_salt
	ItemIndex         string         `db:"item_index" json:"item_index" structs:"item_index"`                            // item_index

}

// Vote represents a row from 'public.votes'.
type Vote struct {
	ID                int64     `db:"id" json:"id" structs:"-"`                                                  // id
	TransferIndex     []byte    `db:"transfer_index" json:"transfer_index" structs:"transfer_index"`             // transfer_index
	Choice            int       `db:"choice" json:"choice" structs:"choice"`                                     // choice
	RarimoTransaction []byte    `db:"rarimo_transaction" json:"rarimo_transaction" structs:"rarimo_transaction"` // rarimo_transaction
	CreatedAt         time.Time `db:"created_at" json:"created_at" structs:"created_at"`                         // created_at

}

// Withdrawal represents a row from 'public.withdrawals'.
type Withdrawal struct {
	Origin    []byte         `db:"origin" json:"origin" structs:"-"`                  // origin
	Hash      sql.NullString `db:"hash" json:"hash" structs:"hash"`                   // hash
	Success   sql.NullBool   `db:"success" json:"success" structs:"success"`          // success
	CreatedAt time.Time      `db:"created_at" json:"created_at" structs:"created_at"` // created_at

}
