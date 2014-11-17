package main

type Catalog struct {
	Name    string
	Entries []Item
}

type Item struct {
	ID  int64
	OID int64
	Ra  float64
	Dec float64
}
