package data

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"gitlab.com/distributed_lab/logan/v3/errors"
)

type AccountID string    // {networkName}:{accountAddress}
type PublicKey string    // {networkName}:{publicKey}
type WithdrawalID string // {networkName}:{withdrawalTransactionHash}

func FormatPublicKey(network string, publicKey []byte) PublicKey {
	return PublicKey(fmt.Sprintf("%s:%s", network, hex.EncodeToString(publicKey)))
}

func FormatAccountID(network string, accountID []byte) AccountID {
	return AccountID(fmt.Sprintf("%s:%s", network, hex.EncodeToString(accountID)))
}

func FormatWithdrawalID(network string, withdrawalID string) WithdrawalID {
	return WithdrawalID(fmt.Sprintf("%s:%s", network, withdrawalID))
}

func DecodePublicKey(raw PublicKey) (string, []byte, error) {
	net, keys, err := decodeNetAndIdentifiers(string(raw), ndecoder(1))
	return net, keys[0], err
}

func DecodeAccountID(raw AccountID) (string, []byte, error) {
	net, keys, err := decodeNetAndIdentifiers(string(raw), ndecoder(1))
	return net, keys[0], err
}

func DecodeWithdrawalID(raw WithdrawalID) (string, string, error) {
	if raw == "" {
		return "", "", errors.New("empty")
	}

	network, withdrawalID, ok := strings.Cut(string(raw), ":")
	if !ok {
		return "", "", fmt.Errorf("unexpected identifier format: %s, should be {network}:{identifier}", raw)
	}

	return network, withdrawalID, nil
}

func (id AccountID) String() string {
	return string(id)
}

func (pk PublicKey) String() string {
	return string(pk)
}

func (id WithdrawalID) String() string {
	return string(id)
}

func (id WithdrawalID) Bytes() []byte {
	return []byte(id)
}

func decodeNetAndIdentifiers(raw string, decodefn func(raw string) ([][]byte, error)) (string, [][]byte, error) {
	if raw == "" {
		return "", nil, errors.New("empty")
	}

	network, identifiers, ok := strings.Cut(raw, ":")
	if !ok {
		return "", nil, fmt.Errorf("unexpected identifier format: %s, should be {network}:{identifier}", raw)
	}

	decoded, err := decodefn(identifiers)
	if err != nil {
		return "", nil, err
	}

	return network, decoded, nil
}

func decode(raw string) ([][]byte, error) {
	decoded, err := hexutil.Decode(raw)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode hex string")
	}

	return [][]byte{decoded}, nil
}

func ndecoder(n int) func(raw string) ([][]byte, error) {
	return func(raw string) ([][]byte, error) {
		split := strings.Split(raw, ":")
		if len(split) != n {
			return nil, fmt.Errorf("unexpected identifier format: %s, should have %d identifiers divided by :, e.g. {foo}:{bar} in case n == 2", raw, n)
		}

		decoded := make([][]byte, n)

		for i := 0; i < n; i++ {
			ident, err := decode(raw)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("failed to decode identifier %s from hex", split[i]))
			}

			decoded[i] = ident[0]
		}

		return decoded, nil

	}
}
