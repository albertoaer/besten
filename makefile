main: build_besten build_symdump

build_besten:
	go build ./cmd/besten

build_symdump:
	go build ./cmd/symdump