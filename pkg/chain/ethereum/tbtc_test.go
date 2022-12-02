package ethereum

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/keep-network/keep-core/pkg/internal/testutils"
	"github.com/keep-network/keep-core/pkg/protocol/group"
)

func TestConvertSignaturesToChainFormat(t *testing.T) {
	signatureSize := 65

	signature1 := common.LeftPadBytes([]byte{1, 2, 3}, signatureSize)
	signature2 := common.LeftPadBytes([]byte{4, 5, 6}, signatureSize)
	signature3 := common.LeftPadBytes([]byte{7}, signatureSize)
	signature4 := common.LeftPadBytes([]byte{8, 9, 10}, signatureSize)
	signature5 := common.LeftPadBytes([]byte{11, 12, 13}, signatureSize)

	invalidSignature := common.LeftPadBytes([]byte("invalid"), signatureSize-1)

	var tests = map[string]struct {
		signaturesMap map[group.MemberIndex][]byte
		expectedError error
	}{
		"one valid signature": {
			signaturesMap: map[uint8][]byte{
				1: signature1,
			},
		},
		"five valid signatures": {
			signaturesMap: map[group.MemberIndex][]byte{
				3: signature3,
				1: signature1,
				4: signature4,
				5: signature5,
				2: signature2,
			},
		},
		"invalid signature": {
			signaturesMap: map[group.MemberIndex][]byte{
				1: signature1,
				2: invalidSignature,
			},
			expectedError: fmt.Errorf("invalid signature size for member [2] got [64] bytes but [65] bytes required"),
		},
	}
	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			indicesSlice, signaturesSlice, err :=
				convertSignaturesToChainFormat(test.signaturesMap)

			if !reflect.DeepEqual(err, test.expectedError) {
				t.Errorf(
					"unexpected error\nexpected: %v\nactual:   %v\n",
					test.expectedError,
					err,
				)
			}

			if test.expectedError == nil {
				testutils.AssertIntsEqual(
					t,
					"member indices slice length",
					len(test.signaturesMap),
					len(indicesSlice),
				)

				testutils.AssertIntsEqual(
					t,
					"signatures slice length",
					signatureSize*len(test.signaturesMap),
					len(signaturesSlice),
				)
			}

			for i, actualMemberIndex := range indicesSlice {
				memberIndex := group.MemberIndex(actualMemberIndex.Uint64())

				actualSignature := signaturesSlice[signatureSize*i : signatureSize*(i+1)]
				if !bytes.Equal(
					test.signaturesMap[memberIndex],
					actualSignature,
				) {
					t.Errorf(
						"invalid signatures for member %v\nexpected: %v\nactual:   %v\n",
						actualMemberIndex,
						test.signaturesMap[memberIndex],
						actualSignature,
					)
				}
			}
		})
	}
}

func TestConvertPubKeyToChainFormat(t *testing.T) {
	bytes30 := []byte{229, 19, 136, 216, 125, 157, 135, 142, 67, 130,
		136, 13, 76, 188, 32, 218, 243, 134, 95, 73, 155, 24, 38, 73, 117, 90,
		215, 95, 216, 19}
	bytes31 := []byte{182, 142, 176, 51, 131, 130, 111, 197, 191, 103, 180, 137,
		171, 101, 34, 78, 251, 234, 118, 184, 16, 116, 238, 82, 131, 153, 134,
		17, 46, 158, 94}

	expectedResult := [64]byte{
		// padding
		00, 00,
		// bytes30
		229, 19, 136, 216, 125, 157, 135, 142, 67, 130, 136, 13, 76, 188, 32,
		218, 243, 134, 95, 73, 155, 24, 38, 73, 117, 90, 215, 95, 216, 19,
		// padding
		00,
		// bytes31
		182, 142, 176, 51, 131, 130, 111, 197, 191, 103, 180, 137, 171, 101, 34,
		78, 251, 234, 118, 184, 16, 116, 238, 82, 131, 153, 134, 17, 46, 158, 94,
	}

	actualResult, err := convertPubKeyToChainFormat(
		&ecdsa.PublicKey{
			X: new(big.Int).SetBytes(bytes30),
			Y: new(big.Int).SetBytes(bytes31),
		},
	)

	if err != nil {
		t.Errorf("unexpected error [%v]", err)
	}

	testutils.AssertBytesEqual(
		t,
		expectedResult[:],
		actualResult[:],
	)
}
