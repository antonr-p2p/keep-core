package signing

import (
	"context"
	"fmt"
	"github.com/bnb-chain/tss-lib/tss"
	"github.com/keep-network/keep-core/pkg/crypto/ephemeral"
	"github.com/keep-network/keep-core/pkg/internal/tecdsatest"
	"github.com/keep-network/keep-core/pkg/internal/testutils"
	"github.com/keep-network/keep-core/pkg/protocol/group"
	"github.com/keep-network/keep-core/pkg/tecdsa"
	"math/big"
	"reflect"
	"strings"
	"testing"
	"time"
)

// TODO: This file contains unit tests that stress each protocol phase
//       separately. We should also develop integration tests checking the
//       whole signing protocol.

const (
	groupSize          = 3
	dishonestThreshold = 0
	sessionID          = "session-1"
)

func TestGenerateEphemeralKeyPair(t *testing.T) {
	members, err := initializeEphemeralKeyPairGeneratingMembersGroup(
		dishonestThreshold,
		groupSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Generate ephemeral key pairs for each group member.
	messages := make(map[group.MemberIndex]*ephemeralPublicKeyMessage)
	for _, member := range members {
		message, err := member.generateEphemeralKeyPair()
		if err != nil {
			t.Fatal(err)
		}
		messages[member.id] = message
	}

	// Assert that each member has a correct state.
	for _, member := range members {
		// Assert the right key pairs count is stored in the member's state.
		expectedKeyPairsCount := groupSize - 1
		actualKeyPairsCount := len(member.ephemeralKeyPairs)
		testutils.AssertIntsEqual(
			t,
			fmt.Sprintf(
				"number of stored ephemeral key pairs for member [%v]",
				member.id,
			),
			expectedKeyPairsCount,
			actualKeyPairsCount,
		)

		// Assert the member does not hold a key pair with itself.
		_, ok := member.ephemeralKeyPairs[member.id]
		if ok {
			t.Errorf(
				"[member:%v] found ephemeral key pair generated to self",
				member.id,
			)
		}

		// Assert key pairs are non-nil.
		for otherMemberID, keyPair := range member.ephemeralKeyPairs {
			if keyPair == nil {
				t.Errorf(
					"[member:%v] key pair not set for member [%v]",
					member.id,
					otherMemberID,
				)
			}

			if keyPair.PrivateKey == nil {
				t.Errorf(
					"[member:%v] key pair's private key not set for member [%v]",
					member.id,
					otherMemberID,
				)
			}

			if keyPair.PublicKey == nil {
				t.Errorf(
					"[member:%v] key pair's public key not set for member [%v]",
					member.id,
					otherMemberID,
				)
			}
		}
	}

	// Assert that each message is formed correctly.
	for memberID, message := range messages {
		// We should always be the sender of our own messages.
		testutils.AssertIntsEqual(
			t,
			"message sender",
			int(memberID),
			int(message.senderID),
		)

		testutils.AssertIntsEqual(
			t,
			"ephemeral public keys count",
			groupSize-1,
			len(message.ephemeralPublicKeys),
		)

		// We should not generate an ephemeral key for ourselves.
		_, ok := message.ephemeralPublicKeys[memberID]
		if ok {
			t.Errorf("found ephemeral key generated to self")
		}

		// We should always use the proper session ID.
		testutils.AssertStringsEqual(
			t,
			fmt.Sprintf(
				"session ID in message generated by member [%v]",
				memberID,
			),
			sessionID,
			message.sessionID,
		)
	}
}

func TestGenerateSymmetricKeys(t *testing.T) {
	members, messages, err := initializeSymmetricKeyGeneratingMembersGroup(
		dishonestThreshold,
		groupSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Generate symmetric keys for each group member.
	for _, member := range members {
		var receivedMessages []*ephemeralPublicKeyMessage
		for _, message := range messages {
			if message.senderID != member.id {
				receivedMessages = append(receivedMessages, message)
			}
		}

		err := member.generateSymmetricKeys(receivedMessages)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Assert that each member has a correct state.
	for _, member := range members {
		// Assert the right keys count is stored in the member's state.
		expectedKeysCount := groupSize - 1
		actualKeysCount := len(member.symmetricKeys)
		testutils.AssertIntsEqual(
			t,
			fmt.Sprintf(
				"number of stored symmetric keys for member [%v]",
				member.id,
			),
			expectedKeysCount,
			actualKeysCount,
		)

		// Assert all symmetric keys stored by this member are correct.
		for otherMemberID, actualKey := range member.symmetricKeys {
			var otherMemberEphemeralPublicKey *ephemeral.PublicKey
			for _, message := range messages {
				if message.senderID == otherMemberID {
					if ephemeralPublicKey, ok := message.ephemeralPublicKeys[member.id]; ok {
						otherMemberEphemeralPublicKey = ephemeralPublicKey
					}
				}
			}

			if otherMemberEphemeralPublicKey == nil {
				t.Errorf(
					"[member:%v] no ephemeral public key from member [%v]",
					member.id,
					otherMemberID,
				)
			}

			expectedKey := ephemeral.SymmetricKey(
				member.ephemeralKeyPairs[otherMemberID].PrivateKey.Ecdh(
					otherMemberEphemeralPublicKey,
				),
			)

			if !reflect.DeepEqual(
				expectedKey,
				actualKey,
			) {
				t.Errorf(
					"[member:%v] wrong symmetric key for member [%v]",
					member.id,
					otherMemberID,
				)
			}
		}
	}
}

func TestGenerateSymmetricKeys_InvalidEphemeralPublicKeyMessage(t *testing.T) {
	members, messages, err := initializeSymmetricKeyGeneratingMembersGroup(
		dishonestThreshold,
		groupSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Corrupt the message sent by member 2 by removing the ephemeral
	// public key generated for member 3.
	misbehavingMemberID := group.MemberIndex(2)
	delete(messages[misbehavingMemberID-1].ephemeralPublicKeys, 3)

	// Generate symmetric keys for each group member.
	for _, member := range members {
		var receivedMessages []*ephemeralPublicKeyMessage
		for _, message := range messages {
			if message.senderID != member.id {
				receivedMessages = append(receivedMessages, message)
			}
		}

		err := member.generateSymmetricKeys(receivedMessages)

		var expectedErr error
		// The misbehaved member should not get an error.
		if member.id != misbehavingMemberID {
			expectedErr = fmt.Errorf(
				"member [%v] sent invalid ephemeral "+
					"public key message",
				misbehavingMemberID,
			)
		}

		if !reflect.DeepEqual(expectedErr, err) {
			t.Errorf(
				"unexpected error\nexpected: %v\nactual:   %v\n",
				expectedErr,
				err,
			)
		}
	}
}

func TestTssRoundOne(t *testing.T) {
	members, err := initializeTssRoundOneMembersGroup(
		dishonestThreshold,
		groupSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Perform TSS round one for each group member.
	messages := make(map[group.MemberIndex]*tssRoundOneMessage)
	for _, member := range members {
		ctx, cancelCtx := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)

		tssRoundOneMessage, err := member.tssRoundOne(ctx)
		if err != nil {
			cancelCtx()
			t.Fatal(err)
		}
		messages[member.id] = tssRoundOneMessage

		cancelCtx()
	}

	// Assert that each member has a correct state.
	for _, member := range members {
		if !strings.Contains(member.tssParty.String(), "round: 1") {
			t.Errorf("wrong round number for member [%v]", member.id)
		}
	}

	// Assert that each message is formed correctly.
	for memberID, message := range messages {
		assertOutgoingAggregateMessage(
			t,
			*message,
			memberID,
			members[memberID-1].symmetricKeys,
		)
	}
}

func TestTssRoundOne_OutgoingMessageTimeout(t *testing.T) {
	members, err := initializeTssRoundOneMembersGroup(
		dishonestThreshold,
		groupSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Perform TSS round one for each group member.
	for _, member := range members {
		// To simulate the outgoing message timeout we do two things:
		// - we pass an already cancelled context
		// - we make sure no message is emitted from the channel by overwriting
		//   the existing channel with a new one that won't receive any
		//   messages from the underlying TSS local party
		ctx, cancelCtx := context.WithCancel(context.Background())
		cancelCtx()
		member.tssOutgoingMessagesChan = make(<-chan tss.Message)

		_, err := member.tssRoundOne(ctx)

		expectedErr := fmt.Errorf(
			"TSS round one outgoing messages were not generated on time",
		)
		if !reflect.DeepEqual(expectedErr, err) {
			t.Errorf(
				"unexpected error for member [%v]\n"+
					"expected: %v\n"+
					"actual:   %v\n",
				member.id,
				expectedErr,
				err,
			)
		}
	}
}

func TestTssRoundOne_SymmetricKeyMissing(t *testing.T) {
	members, err := initializeTssRoundOneMembersGroup(
		dishonestThreshold,
		groupSize,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Perform TSS round one for each group member.
	for _, member := range members {
		// Cleanup symmetric key cache.
		member.symmetricKeys = make(map[group.MemberIndex]ephemeral.SymmetricKey)

		ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Second)

		_, err := member.tssRoundOne(ctx)

		if !strings.Contains(
			err.Error(),
			"cannot get symmetric key with member",
		) {
			t.Errorf("wrong error for member [%v]: [%v]", member.id, err)
		}

		cancelCtx()
	}
}

func assertOutgoingAggregateMessage(
	t *testing.T,
	message struct{
		senderID group.MemberIndex
		broadcastPayload []byte
	    peersPayload map[group.MemberIndex][]byte
		sessionID string
	},
	memberID group.MemberIndex,
	memberSymmetricKeys map[group.MemberIndex]ephemeral.SymmetricKey,
) {
	// We should always be the sender of our own messages.
	testutils.AssertIntsEqual(
		t,
		fmt.Sprintf(
			"message sender in message generated by member [%v]",
			memberID,
		),
		int(memberID),
		int(message.senderID),
	)

	// We should always generate a broadcast payload.
	if len(message.broadcastPayload) == 0 {
		t.Errorf(
			"empty broadcast payload in message generated by member [%v]",
			memberID,
		)
	}

	// We should always generate groupSize-1 of peers payloads.
	testutils.AssertIntsEqual(
		t,
		fmt.Sprintf(
			"count of peers payloads in message "+
				"generated by member [%v]",
			memberID,
		),
		groupSize-1,
		len(message.peersPayload),
	)

	// Each P2P payload should be encrypted using the proper symmetric key.
	for receiverID, encryptedPayload := range message.peersPayload {
		symmetricKey := memberSymmetricKeys[receiverID]
		if _, err := symmetricKey.Decrypt(encryptedPayload); err != nil {
			t.Errorf(
				"payload for member [%v] in message generated "+
					"by member [%v] is encrypted using "+
					"the wrong symmetric key: [%v]",
				receiverID,
				memberID,
				err,
			)
		}
	}

	// We should always use the proper session ID.
	testutils.AssertStringsEqual(
		t,
		fmt.Sprintf(
			"session ID in message generated by member [%v]",
			memberID,
		),
		sessionID,
		message.sessionID,
	)
}

func initializeEphemeralKeyPairGeneratingMembersGroup(
	dishonestThreshold int,
	groupSize int,
) ([]*ephemeralKeyPairGeneratingMember, error) {
	signingGroup := group.NewGroup(dishonestThreshold, groupSize)

	testData, err := tecdsatest.LoadPrivateKeyShareTestFixtures(groupSize)
	if err != nil {
		return nil, fmt.Errorf("failed to load test data: [%v]", err)
	}

	var members []*ephemeralKeyPairGeneratingMember
	for i := 1; i <= groupSize; i++ {
		id := group.MemberIndex(i)

		members = append(members, &ephemeralKeyPairGeneratingMember{
			member: &member{
				logger:          &testutils.MockLogger{},
				id:              id,
				group:           signingGroup,
				sessionID:       sessionID,
				message:         big.NewInt(100),
				privateKeyShare: tecdsa.NewPrivateKeyShare(testData[i-1]),
			},
			ephemeralKeyPairs: make(map[group.MemberIndex]*ephemeral.KeyPair),
		})
	}

	return members, nil
}
func initializeSymmetricKeyGeneratingMembersGroup(
	dishonestThreshold int,
	groupSize int,
) (
	[]*symmetricKeyGeneratingMember,
	[]*ephemeralPublicKeyMessage,
	error,
) {
	var symmetricKeyGeneratingMembers []*symmetricKeyGeneratingMember
	var ephemeralPublicKeyMessages []*ephemeralPublicKeyMessage

	ephemeralKeyPairGeneratingMembers, err :=
		initializeEphemeralKeyPairGeneratingMembersGroup(
			dishonestThreshold,
			groupSize,
		)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"cannot generate ephemeral key pair generating "+
				"members group: [%v]",
			err,
		)
	}

	for _, member := range ephemeralKeyPairGeneratingMembers {
		message, err := member.generateEphemeralKeyPair()
		if err != nil {
			return nil, nil, fmt.Errorf(
				"cannot generate ephemeral key pair for member [%v]: [%v]",
				member.id,
				err,
			)
		}

		symmetricKeyGeneratingMembers = append(
			symmetricKeyGeneratingMembers,
			member.initializeSymmetricKeyGeneration(),
		)
		ephemeralPublicKeyMessages = append(ephemeralPublicKeyMessages, message)
	}

	return symmetricKeyGeneratingMembers, ephemeralPublicKeyMessages, nil
}

func initializeTssRoundOneMembersGroup(
	dishonestThreshold int,
	groupSize int,
) (
	[]*tssRoundOneMember,
	error,
) {
	var tssRoundOneMembers []*tssRoundOneMember

	symmetricKeyGeneratingMembers, ephemeralPublicKeyMessages, err :=
		initializeSymmetricKeyGeneratingMembersGroup(
			dishonestThreshold,
			groupSize,
		)
	if err != nil {
		return nil, fmt.Errorf(
			"cannot generate symmetric key generating members group: [%v]",
			err,
		)
	}

	for _, member := range symmetricKeyGeneratingMembers {
		var receivedMessages []*ephemeralPublicKeyMessage
		for _, message := range ephemeralPublicKeyMessages {
			if message.senderID != member.id {
				receivedMessages = append(receivedMessages, message)
			}
		}

		err := member.generateSymmetricKeys(receivedMessages)
		if err != nil {
			return nil, fmt.Errorf(
				"cannot generate symmetric keys for member [%v]: [%v]",
				member.id,
				err,
			)
		}

		tssRoundOneMembers = append(
			tssRoundOneMembers,
			member.initializeTssRoundOne(),
		)
	}

	return tssRoundOneMembers, nil
}