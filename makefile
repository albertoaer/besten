main: build_besten build_bbctester

build_besten:
	go build ./cmd/besten

build_bbctester:
	go build ./cmd/bbctester