ARCH=6 #6g, 6l, 5g, 5l...
SRC=buddha.go
OUT=buddha

all:
	6g -o $(OUT).$(ARCH) $(SRC)
	6l -o $(OUT) $(OUT).$(ARCH)

run: all
	./$(OUT)

fmt:
	gofmt -w=true $(SRC)

clean:
	rm *.$(ARCH)

mrpoprer: clean
	rm $(OUT)

