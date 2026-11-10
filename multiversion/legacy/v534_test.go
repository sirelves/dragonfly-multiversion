package legacy

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestOldFormResponses(t *testing.T) {
	closed := delta534.toLatest(534, &ModalFormResponse534{FormID: 3, ResponseData: []byte("null\n")})[0].(*packet.ModalFormResponse)
	if _, ok := closed.ResponseData.Value(); ok {
		t.Fatal("a closed form carries response data")
	}
	if r, ok := closed.CancelReason.Value(); !ok || r != packet.ModalFormCancelReasonUserClosed {
		t.Fatal("a closed form has no cancel reason")
	}
	answered := delta534.toLatest(534, &ModalFormResponse534{FormID: 3, ResponseData: []byte("[1,\"a\"]")})[0].(*packet.ModalFormResponse)
	if data, ok := answered.ResponseData.Value(); !ok || string(data) != "[1,\"a\"]" || answered.FormID != 3 {
		t.Fatal("form answer lost")
	}
	if _, ok := answered.CancelReason.Value(); ok {
		t.Fatal("an answered form is cancelled")
	}
}
