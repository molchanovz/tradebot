package tradeplus

import (
	"fmt"
	"testing"
)

func TestSetArticleDescription(t *testing.T) {
	got := SetArticleDescription("SM-DH-10L")
	fmt.Println(*got)
}
