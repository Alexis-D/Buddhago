SRC=buddha.go
OBJ=buddha.6
OUT=buddha

all:
	6g -o $(OBJ) $(SRC)
	6l -o $(OUT) $(OBJ)

run: all
	./$(OUT)

fmt:
	gofmt -w=true $(SRC)

clean:
	rm *.6

mrpoprer: clean
	rm $(OUT)

