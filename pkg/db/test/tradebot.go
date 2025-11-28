package test

import (
	"testing"
	"time"

	"tradebot/pkg/db"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/go-pg/pg/v10/orm"
)

type CabinetOpFunc func(t *testing.T, dbo orm.DB, in *db.Cabinet) Cleaner

func Cabinet(t *testing.T, dbo orm.DB, in *db.Cabinet, ops ...CabinetOpFunc) (*db.Cabinet, Cleaner) {
	repo := db.NewTradebotRepo(dbo)
	var cleaners []Cleaner

	// Fill the incoming entity
	if in == nil {
		in = &db.Cabinet{}
	}

	// Check if PKs are provided
	if in.ID != 0 {
		// Fetch the entity by PK
		cabinet, err := repo.CabinetByID(t.Context(), in.ID, repo.FullCabinet())
		if err != nil {
			t.Fatal(err)
		}

		// We must find the entity by PK
		if cabinet == nil {
			t.Fatalf("the entity Cabinet is not found by provided PKs ID=%v", in.ID)
		}

		// Return if found without real cleanup
		return cabinet, emptyClean
	}

	for _, op := range ops {
		if cl := op(t, dbo, in); cl != nil {
			cleaners = append(cleaners, cl)
		}
	}

	// Create the main entity
	cabinet, err := repo.AddCabinet(t.Context(), in)
	if err != nil {
		t.Fatal(err)
	}

	return cabinet, func() {
		if _, err := dbo.ModelContext(t.Context(), &db.Cabinet{ID: cabinet.ID}).WherePK().Delete(); err != nil {
			t.Fatal(err)
		}

		// Clean up related entities from the last to the first
		for i := len(cleaners) - 1; i >= 0; i-- {
			cleaners[i]()
		}
	}
}

func WithFakeCabinet(t *testing.T, dbo orm.DB, in *db.Cabinet) Cleaner {
	if in.Name == "" {
		in.Name = cutS(gofakeit.Sentence(6), 64)
	}

	if in.Key == "" {
		in.Key = cutS(gofakeit.Sentence(10), 1024)
	}

	if in.Marketplace == "" {
		in.Marketplace = cutS(gofakeit.Sentence(10), 0)
	}

	if in.Type == "" {
		in.Type = cutS(gofakeit.Sentence(10), 0)
	}

	if in.StatusID == 0 {
		in.StatusID = 1
	}

	return emptyClean
}

type OrderOpFunc func(t *testing.T, dbo orm.DB, in *db.Order) Cleaner

func Order(t *testing.T, dbo orm.DB, in *db.Order, ops ...OrderOpFunc) (*db.Order, Cleaner) {
	repo := db.NewTradebotRepo(dbo)
	var cleaners []Cleaner

	// Fill the incoming entity
	if in == nil {
		in = &db.Order{}
	}

	// Check if PKs are provided
	if in.ID != 0 {
		// Fetch the entity by PK
		order, err := repo.OrderByID(t.Context(), in.ID, repo.FullOrder())
		if err != nil {
			t.Fatal(err)
		}

		// We must find the entity by PK
		if order == nil {
			t.Fatalf("the entity Order is not found by provided PKs ID=%v", in.ID)
		}

		// Return if found without real cleanup
		return order, emptyClean
	}

	for _, op := range ops {
		if cl := op(t, dbo, in); cl != nil {
			cleaners = append(cleaners, cl)
		}
	}

	// Create the main entity
	order, err := repo.AddOrder(t.Context(), in)
	if err != nil {
		t.Fatal(err)
	}

	return order, func() {
		if _, err := dbo.ModelContext(t.Context(), &db.Order{ID: order.ID}).WherePK().Delete(); err != nil {
			t.Fatal(err)
		}

		// Clean up related entities from the last to the first
		for i := len(cleaners) - 1; i >= 0; i-- {
			cleaners[i]()
		}
	}
}

func WithOrderRelations(t *testing.T, dbo orm.DB, in *db.Order) Cleaner {
	var cleaners []Cleaner

	// Prepare main relations
	if in.Cabinet == nil {
		in.Cabinet = &db.Cabinet{}
	}

	// Check embedded entities by FK

	// Cabinet. Check if all FKs are provided.

	if in.CabinetID != 0 {
		in.Cabinet.ID = in.CabinetID
	}

	// Fetch the relation. It creates if the FKs are provided it fetch from DB by PKs. Else it creates new one.
	{
		rel, relatedCleaner := Cabinet(t, dbo, in.Cabinet, WithFakeCabinet)
		in.Cabinet = rel
		in.CabinetID = rel.ID

		cleaners = append(cleaners, relatedCleaner)
	}

	return func() {
		// Clean up related entities from the last to the first
		for i := len(cleaners) - 1; i >= 0; i-- {
			cleaners[i]()
		}
	}
}

func WithFakeOrder(t *testing.T, dbo orm.DB, in *db.Order) Cleaner {
	if in.PostingNumber == "" {
		in.PostingNumber = cutS(gofakeit.Sentence(3), 32)
	}

	if in.Article == "" {
		in.Article = cutS(gofakeit.Sentence(10), 128)
	}

	if in.Count == 0 {
		in.Count = gofakeit.IntRange(1, 10)
	}

	if in.CabinetID == 0 {
		in.CabinetID = gofakeit.IntRange(1, 10)
	}

	if in.CreatedAt.IsZero() {
		in.CreatedAt = time.Now()
	}

	if in.StatusID == 0 {
		in.StatusID = 1
	}

	return emptyClean
}

type StockOpFunc func(t *testing.T, dbo orm.DB, in *db.Stock) Cleaner

func Stock(t *testing.T, dbo orm.DB, in *db.Stock, ops ...StockOpFunc) (*db.Stock, Cleaner) {
	repo := db.NewTradebotRepo(dbo)
	var cleaners []Cleaner

	// Fill the incoming entity
	if in == nil {
		in = &db.Stock{}
	}

	// Check if PKs are provided
	if in.ID != 0 {
		// Fetch the entity by PK
		stock, err := repo.StockByID(t.Context(), in.ID, repo.FullStock())
		if err != nil {
			t.Fatal(err)
		}

		// We must find the entity by PK
		if stock == nil {
			t.Fatalf("the entity Stock is not found by provided PKs ID=%v", in.ID)
		}

		// Return if found without real cleanup
		return stock, emptyClean
	}

	for _, op := range ops {
		if cl := op(t, dbo, in); cl != nil {
			cleaners = append(cleaners, cl)
		}
	}

	// Create the main entity
	stock, err := repo.AddStock(t.Context(), in)
	if err != nil {
		t.Fatal(err)
	}

	return stock, func() {
		if _, err := dbo.ModelContext(t.Context(), &db.Stock{ID: stock.ID}).WherePK().Delete(); err != nil {
			t.Fatal(err)
		}

		// Clean up related entities from the last to the first
		for i := len(cleaners) - 1; i >= 0; i-- {
			cleaners[i]()
		}
	}
}

func WithStockRelations(t *testing.T, dbo orm.DB, in *db.Stock) Cleaner {
	var cleaners []Cleaner

	// Prepare main relations
	if in.Cabinet == nil {
		in.Cabinet = &db.Cabinet{}
	}

	// Check embedded entities by FK

	// Cabinet. Check if all FKs are provided.

	if in.CabinetID != 0 {
		in.Cabinet.ID = in.CabinetID
	}

	// Fetch the relation. It creates if the FKs are provided it fetch from DB by PKs. Else it creates new one.
	{
		rel, relatedCleaner := Cabinet(t, dbo, in.Cabinet, WithFakeCabinet)
		in.Cabinet = rel
		in.CabinetID = rel.ID

		cleaners = append(cleaners, relatedCleaner)
	}

	return func() {
		// Clean up related entities from the last to the first
		for i := len(cleaners) - 1; i >= 0; i-- {
			cleaners[i]()
		}
	}
}

func WithFakeStock(t *testing.T, dbo orm.DB, in *db.Stock) Cleaner {
	if in.Article == "" {
		in.Article = cutS(gofakeit.Sentence(6), 64)
	}

	if in.UpdatedAt.IsZero() {
		in.UpdatedAt = gofakeit.DateRange(time.Now().Add(5*time.Minute), time.Now().Add(1*time.Hour))
	}

	if in.CabinetID == 0 {
		in.CabinetID = gofakeit.IntRange(1, 10)
	}

	return emptyClean
}

type UserOpFunc func(t *testing.T, dbo orm.DB, in *db.User) Cleaner

func User(t *testing.T, dbo orm.DB, in *db.User, ops ...UserOpFunc) (*db.User, Cleaner) {
	repo := db.NewTradebotRepo(dbo)
	var cleaners []Cleaner

	// Fill the incoming entity
	if in == nil {
		in = &db.User{}
	}

	// Check if PKs are provided
	if in.ID != 0 {
		// Fetch the entity by PK
		user, err := repo.UserByID(t.Context(), in.ID, repo.FullUser())
		if err != nil {
			t.Fatal(err)
		}

		// We must find the entity by PK
		if user == nil {
			t.Fatalf("the entity User is not found by provided PKs ID=%v", in.ID)
		}

		// Return if found without real cleanup
		return user, emptyClean
	}

	for _, op := range ops {
		if cl := op(t, dbo, in); cl != nil {
			cleaners = append(cleaners, cl)
		}
	}

	// Create the main entity
	user, err := repo.AddUser(t.Context(), in)
	if err != nil {
		t.Fatal(err)
	}

	return user, func() {
		if _, err := dbo.ModelContext(t.Context(), &db.User{ID: user.ID}).WherePK().Delete(); err != nil {
			t.Fatal(err)
		}

		// Clean up related entities from the last to the first
		for i := len(cleaners) - 1; i >= 0; i-- {
			cleaners[i]()
		}
	}
}

func WithFakeUser(t *testing.T, dbo orm.DB, in *db.User) Cleaner {
	if in.TgID == 0 {
		in.TgID = int64(gofakeit.IntRange(1, 10))
	}

	if in.IsAdmin == false {
		in.IsAdmin = gofakeit.Bool()
	}

	if in.StatusID == 0 {
		in.StatusID = 1
	}

	if in.CreatedAt.IsZero() {
		in.CreatedAt = time.Now()
	}

	return emptyClean
}

type ReviewOpFunc func(t *testing.T, dbo orm.DB, in *db.Review) Cleaner

func Review(t *testing.T, dbo orm.DB, in *db.Review, ops ...ReviewOpFunc) (*db.Review, Cleaner) {
	repo := db.NewTradebotRepo(dbo)
	var cleaners []Cleaner

	// Fill the incoming entity
	if in == nil {
		in = &db.Review{}
	}

	// Check if PKs are provided
	if in.ID != 0 {
		// Fetch the entity by PK
		review, err := repo.ReviewByID(t.Context(), in.ID, repo.FullReview())
		if err != nil {
			t.Fatal(err)
		}

		// We must find the entity by PK
		if review == nil {
			t.Fatalf("the entity Review is not found by provided PKs ID=%v", in.ID)
		}

		// Return if found without real cleanup
		return review, emptyClean
	}

	for _, op := range ops {
		if cl := op(t, dbo, in); cl != nil {
			cleaners = append(cleaners, cl)
		}
	}

	// Create the main entity
	review, err := repo.AddReview(t.Context(), in)
	if err != nil {
		t.Fatal(err)
	}

	return review, func() {
		if _, err := dbo.ModelContext(t.Context(), &db.Review{ID: review.ID}).WherePK().Delete(); err != nil {
			t.Fatal(err)
		}

		// Clean up related entities from the last to the first
		for i := len(cleaners) - 1; i >= 0; i-- {
			cleaners[i]()
		}
	}
}

func WithReviewRelations(t *testing.T, dbo orm.DB, in *db.Review) Cleaner {
	var cleaners []Cleaner

	// Prepare main relations
	if in.Cabinet == nil {
		in.Cabinet = &db.Cabinet{}
	}

	// Check embedded entities by FK

	// Cabinet. Check if all FKs are provided.

	if in.CabinetID != 0 {
		in.Cabinet.ID = in.CabinetID
	}

	// Fetch the relation. It creates if the FKs are provided it fetch from DB by PKs. Else it creates new one.
	{
		rel, relatedCleaner := Cabinet(t, dbo, in.Cabinet, WithFakeCabinet)
		in.Cabinet = rel
		in.CabinetID = rel.ID

		cleaners = append(cleaners, relatedCleaner)
	}

	return func() {
		// Clean up related entities from the last to the first
		for i := len(cleaners) - 1; i >= 0; i-- {
			cleaners[i]()
		}
	}
}

func WithFakeReview(t *testing.T, dbo orm.DB, in *db.Review) Cleaner {
	if in.CabinetID == 0 {
		in.CabinetID = gofakeit.IntRange(1, 10)
	}

	if in.ExternalID == "" {
		in.ExternalID = cutS(gofakeit.Sentence(10), 128)
	}

	if in.Valuation == 0 {
		in.Valuation = gofakeit.IntRange(1, 10)
	}

	if in.Article == "" {
		in.Article = cutS(gofakeit.Sentence(10), 128)
	}

	if in.CreatedAt.IsZero() {
		in.CreatedAt = time.Now()
	}

	if in.StatusID == 0 {
		in.StatusID = 1
	}

	return emptyClean
}
