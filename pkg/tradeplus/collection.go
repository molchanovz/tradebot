package tradeplus

//go:generate colgen
//colgen:Card,Review
//colgen:Card:Index(NmID)
//colgen:Review:MapP(db),UniqueExternalID,Index(ExternalID)

// MapP converts slice of type T to slice of type M with given converter with pointers.
func MapP[T, M any](a []T, f func(*T) *M) []M {
	n := make([]M, len(a))
	for i := range a {
		n[i] = *f(&a[i])
	}
	return n
}
