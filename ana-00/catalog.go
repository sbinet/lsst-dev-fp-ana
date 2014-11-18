package main

type Catalog struct {
	Name    string
	Entries []Data
}

type SdssData struct {
	ID  int64   `fits:"id"`
	Ra  float64 `fits:"ra"`
	Dec float64 `fits:"dec"`
}

type LsstData struct {
	ID    int64      `fits:"id"`
	Coord [2]float64 `fits:"coord"`
}

type Data struct {
	ID  int64   `fits:"id"`  // unique id
	Ra  float64 `fits:"ra"`  // right-ascend (degrees)
	Dec float64 `fits:"dec"` // declinaison (degrees)
}
