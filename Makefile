ARCH=6 #6g, 6l, 5g, 5l...
SRC=buddha.go
OUT=buddha

all:
	$(ARCH)g -o $(OUT).$(ARCH) $(SRC)
	$(ARCH)l -o $(OUT) $(OUT).$(ARCH)

run: all
	./$(OUT)

fmt:
	gofmt -w=true $(SRC)

clean:
	rm *.$(ARCH)

mrpoprer: clean
	rm $(OUT)

