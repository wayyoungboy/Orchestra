package a2a

import "testing"

func TestConvertACPToWSSystemMessageKeepsStatusText(t *testing.T) {
	message := &ACPMessage{
		Type:    TypeSystem,
		Content: []byte(`{"type":"system","message":"agent ready","level":"info"}`),
	}

	response := ConvertACPToWS(message)
	if response == nil || response.Type != "status" || response.Status != "agent ready" {
		t.Fatalf("ConvertACPToWS() = %#v, want status with system message", response)
	}
}
