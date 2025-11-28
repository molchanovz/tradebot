package wb

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"tradebot/pkg/db"
	"tradebot/pkg/db/test"
)

func TestReviews(t *testing.T) {
	dbo := test.Setup(t)
	repo := db.NewTradebotRepo(dbo)
	cabinet, err := repo.OneCabinet(t.Context(), &db.CabinetSearch{Marketplace: test.Ptr("WB")})
	if err != nil {
		return
	} else if cabinet == nil {
		return
	}

	Convey("Should get google info", t, func() {
		client := NewClient(cabinet.Key)
		got, err := client.Reviews()
		So(err, ShouldBeNil)
		So(got, ShouldNotBeNil)
	})
}

func TestOrders(t *testing.T) {
	dbo := test.Setup(t)
	repo := db.NewTradebotRepo(dbo)
	cabinet, err := repo.OneCabinet(t.Context(), &db.CabinetSearch{Marketplace: test.Ptr("WB")})
	if err != nil {
		return
	} else if cabinet == nil {
		return
	}

	Convey("Should get google info", t, func() {
		client := NewClient(cabinet.Key)
		got, err := client.GetAllOrders(1, 1)
		So(err, ShouldBeNil)
		So(got, ShouldNotBeNil)
	})
}

func TestClient_AnswerReview(t *testing.T) {

	dbo := test.Setup(t)
	repo := db.NewTradebotRepo(dbo)
	cabinet, err := repo.OneCabinet(t.Context(), &db.CabinetSearch{Marketplace: test.Ptr("WB")})
	if err != nil {
		return
	} else if cabinet == nil {
		return
	}

	Convey("Should get google info", t, func() {
		client := NewClient(cabinet.Key)
		err = client.AnswerReview("", "")
		So(err, ShouldBeNil)
	})
}
