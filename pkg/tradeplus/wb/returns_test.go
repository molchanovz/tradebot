package wb

import (
	"github.com/stretchr/testify/require"
	"testing"
	"tradebot/pkg/db"
	"tradebot/pkg/tradeplus"
	"tradebot/pkg/tradeplus/test"
)

func TestReturnsManager_WriteReturns(t *testing.T) {
	_, err := test.Setup()
	require.NoError(t, err)
	require.NotNil(t, test.Cfg)

	cabinet, err := test.Repo.OneCabinet(t.Context(), &db.CabinetSearch{Marketplace: tradeplus.Ptr("WB")})
	require.NoError(t, err)

	m := NewReturnsManager(cabinet.Key)
	_, err = m.WriteReturns()
}
