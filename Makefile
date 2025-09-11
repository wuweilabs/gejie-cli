.PHONY: build clean

build:
	go build -o gejiec .

clean:
	rm -f gejiec
