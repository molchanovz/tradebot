package chatgptsrv

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestReviewManager_Reviews(t *testing.T) {
	gptSrv := NewClient("http://localhost:8076/int/rpc/", &http.Client{Timeout: time.Second * 30})
	req, err := gptSrv.Chatgpt.Send(t.Context(), "")
	if err != nil {
		t.Log(err)
	}

	fmt.Println(req)
}
