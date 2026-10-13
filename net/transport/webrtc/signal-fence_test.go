package webrtc

// These tests pin the signal generation fence of ingestRemoteSignal. SDP and
// ICE signals carry the SHA-256 of the offer SDP bytes they belong to, and
// material whose digest does not match the active generation is dropped before
// Pion state is touched.

import (
	"bytes"
	"crypto/sha256"
	"testing"
	"time"

	pion_webrtc "github.com/pion/webrtc/v4"
	"github.com/sirupsen/logrus"
)

// testHostCandidate is a host candidate line for fence tests.
const testHostCandidate = "candidate:1 1 udp 2130706431 10.5.0.2 54504 typ host"

// fenceIngest drives ingestRemoteSignal against one session, applying a
// sequence of signals the way the tracker routine does.
type fenceIngest struct {
	// sess is the session under test.
	sess *session
	// xmit receives the signals the session transmits.
	xmit func(*WebRtcSignal)
	// applied counts the candidates that reached the ICE applier.
	applied int
}

// newFenceIngest builds a fence driver over a session on pc in the given role.
func newFenceIngest(pc *pion_webrtc.PeerConnection, offerer bool) *fenceIngest {
	f := &fenceIngest{xmit: func(*WebRtcSignal) {}}
	f.sess = &session{
		t:  &sessionTracker{w: &WebRTC{conf: &Config{}}, offerer: offerer},
		pc: pc,
		remoteICE: remoteICECandidateApplier{
			add: func(pion_webrtc.ICECandidateInit) error {
				f.applied++
				return nil
			},
		},
	}
	return f
}

// ingest applies one received SDP and candidate pair.
func (f *fenceIngest) ingest(sdp *WebRtcSdp, ice *WebRtcIce) error {
	phase := ""
	return f.sess.t.ingestRemoteSignal(
		f.sess,
		sdp,
		ice,
		0,
		f.xmit,
		logrus.NewEntry(logrus.New()),
		&phase,
	)
}

// offerDigest computes the generation identity of an offer's SDP bytes.
func offerDigest(sdp string) []byte {
	sum := sha256.Sum256([]byte(sdp))
	return sum[:]
}

// newLocalOffer returns a peer connection holding an unanswered local offer.
func newLocalOffer(t *testing.T) (*pion_webrtc.PeerConnection, pion_webrtc.SessionDescription) {
	t.Helper()

	// Create the connection and install its local offer.
	pc, err := pion_webrtc.NewPeerConnection(pion_webrtc.Configuration{})
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Cleanup(func() { pc.Close() })
	if _, err := pc.CreateDataChannel(dataChannelID, nil); err != nil {
		t.Fatal(err.Error())
	}
	offer, err := pc.CreateOffer(nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	if err := pc.SetLocalDescription(offer); err != nil {
		t.Fatal(err.Error())
	}
	return pc, offer
}

// newTestIce builds one ICE candidate message tagged with offerID.
func newTestIce(t *testing.T, offerID []byte) *WebRtcIce {
	t.Helper()
	mlineIndex := uint16(0)
	ice, err := NewWebRtcIce(&pion_webrtc.ICECandidateInit{
		Candidate:     testHostCandidate,
		SDPMLineIndex: &mlineIndex,
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	ice.OfferId = offerID
	return ice
}

// newAnswererFence returns an answerer fence driver that has applied a valid
// offer, along with that offer.
func newAnswererFence(t *testing.T) (*fenceIngest, pion_webrtc.SessionDescription) {
	t.Helper()

	// The offer comes from a separate offerer connection.
	_, offer := newLocalOffer(t)
	answerPC, err := pion_webrtc.NewPeerConnection(pion_webrtc.Configuration{})
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Cleanup(func() { answerPC.Close() })

	// Apply the offer.
	f := newFenceIngest(answerPC, false)
	err = f.ingest(&WebRtcSdp{
		SdpType: "offer",
		Sdp:     offer.SDP,
		OfferId: offerDigest(offer.SDP),
	}, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	return f, offer
}

// TestIngestRejectsStaleGenerationAnswer asserts an answer whose offer id does
// not match the pending local offer is dropped before Pion state is touched,
// and the answer that matches is applied.
func TestIngestRejectsStaleGenerationAnswer(t *testing.T) {
	// Build the offerer and the answer the remote peer would return.
	offerPC, offer := newLocalOffer(t)
	answerPC, err := pion_webrtc.NewPeerConnection(pion_webrtc.Configuration{})
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Cleanup(func() { answerPC.Close() })
	if err := answerPC.SetRemoteDescription(offer); err != nil {
		t.Fatal(err.Error())
	}
	answer, err := answerPC.CreateAnswer(nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	f := newFenceIngest(offerPC, true)
	f.sess.pendingOfferID = offerDigest(offer.SDP)

	// An answer for another generation is dropped silently.
	err = f.ingest(&WebRtcSdp{
		SdpType: "answer",
		Sdp:     answer.SDP,
		OfferId: offerDigest("retired-generation-offer"),
	}, nil)
	if err != nil {
		t.Fatalf("stale-generation answer returned %v, want silent drop", err)
	}
	if offerPC.RemoteDescription() != nil {
		t.Fatal("stale-generation answer reached Pion")
	}

	// The answer for the pending generation is applied.
	err = f.ingest(&WebRtcSdp{
		SdpType: "answer",
		Sdp:     answer.SDP,
		OfferId: f.sess.pendingOfferID,
	}, nil)
	if err != nil {
		t.Fatalf("matching answer returned %v", err)
	}
	if offerPC.RemoteDescription() == nil {
		t.Fatal("matching answer was not applied")
	}
}

// TestIngestRejectsRoleViolatingOffer asserts an offerer treats an incoming
// offer as a fatal role violation.
func TestIngestRejectsRoleViolatingOffer(t *testing.T) {
	offerPC, _ := newLocalOffer(t)
	f := newFenceIngest(offerPC, true)
	if err := f.ingest(&WebRtcSdp{SdpType: "offer", Sdp: "v=0\r\n"}, nil); err == nil {
		t.Fatal("role-violating offer was not fatal")
	}
}

// TestIngestFiltersCandidatesByGeneration asserts that once an offer is
// active, a candidate for another generation is dropped and a candidate for
// the active generation reaches the ICE applier.
func TestIngestFiltersCandidatesByGeneration(t *testing.T) {
	f, offer := newAnswererFence(t)

	// A candidate for another generation is dropped silently.
	if err := f.ingest(nil, newTestIce(t, offerDigest("some-other-generation"))); err != nil {
		t.Fatalf("stale-generation candidate returned %v, want silent drop", err)
	}
	if f.applied != 0 {
		t.Fatalf("stale-generation candidate reached the ICE applier: %d applied", f.applied)
	}

	// A candidate for the active generation is applied.
	if err := f.ingest(nil, newTestIce(t, offerDigest(offer.SDP))); err != nil {
		t.Fatal(err.Error())
	}
	if f.applied != 1 {
		t.Fatalf("matching-generation candidate applied %d times, want 1", f.applied)
	}
}

// TestIngestBuffersCandidatesUntilOfferLands asserts a candidate that arrives
// before its offer is buffered and applied once the offer lands.
func TestIngestBuffersCandidatesUntilOfferLands(t *testing.T) {
	// The candidate arrives first and is held back.
	_, offer := newLocalOffer(t)
	answerPC, err := pion_webrtc.NewPeerConnection(pion_webrtc.Configuration{})
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Cleanup(func() { answerPC.Close() })
	f := newFenceIngest(answerPC, false)
	if err := f.ingest(nil, newTestIce(t, offerDigest(offer.SDP))); err != nil {
		t.Fatal(err.Error())
	}
	if f.applied != 0 || len(f.sess.pendingRemoteIce) != 1 {
		t.Fatalf("early candidate: applied=%d buffered=%d, want 0 and 1", f.applied, len(f.sess.pendingRemoteIce))
	}

	// The offer lands and flushes the buffered candidate.
	err = f.ingest(&WebRtcSdp{
		SdpType: "offer",
		Sdp:     offer.SDP,
		OfferId: offerDigest(offer.SDP),
	}, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	if f.applied != 1 || len(f.sess.pendingRemoteIce) != 0 {
		t.Fatalf("after offer: applied=%d buffered=%d, want 1 and 0", f.applied, len(f.sess.pendingRemoteIce))
	}
}

// TestAnswererReplaysRetainedAnswerOnDuplicateOffer asserts that the
// byte-identical offer an answerer already answered gets the retained answer
// replayed with the same generation identity, and leaves Pion state untouched.
func TestAnswererReplaysRetainedAnswerOnDuplicateOffer(t *testing.T) {
	// Capture the answers the session transmits.
	_, offer := newLocalOffer(t)
	answerPC, err := pion_webrtc.NewPeerConnection(pion_webrtc.Configuration{})
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Cleanup(func() { answerPC.Close() })
	gatherComplete := pion_webrtc.GatheringCompletePromise(answerPC)
	f := newFenceIngest(answerPC, false)
	var answers []*WebRtcSdp
	f.xmit = func(sig *WebRtcSignal) {
		answers = append(answers, sig.GetSdp())
	}
	rx := &WebRtcSdp{
		SdpType: "offer",
		Sdp:     offer.SDP,
		OfferId: offerDigest(offer.SDP),
	}

	// The first delivery transmits a fresh answer.
	if err := f.ingest(rx, nil); err != nil {
		t.Fatal(err.Error())
	}
	if len(answers) != 1 || answers[0].GetSdpType() != "answer" {
		t.Fatalf("first offer produced %d signals, want one answer", len(answers))
	}
	select {
	case <-gatherComplete:
	case <-time.After(5 * time.Second):
		t.Fatal("answer ICE gathering did not complete")
	}
	stateBefore := answerPC.SignalingState()
	localBefore := answerPC.LocalDescription().SDP

	// The duplicate replays the retained answer without touching Pion.
	if err := f.ingest(rx, nil); err != nil {
		t.Fatal(err.Error())
	}
	if len(answers) != 2 {
		t.Fatalf("duplicate offer produced %d signals, want one replay", len(answers)-1)
	}
	if answers[1].GetSdp() != answers[0].GetSdp() {
		t.Fatal("replayed answer changed the transmitted SDP bytes")
	}
	if !bytes.Equal(answers[1].GetOfferId(), f.sess.rxOfferID) {
		t.Fatal("replayed answer changed the generation identity")
	}
	if answerPC.SignalingState() != stateBefore || answerPC.LocalDescription().SDP != localBefore {
		t.Fatal("replay mutated the peer connection")
	}
}
