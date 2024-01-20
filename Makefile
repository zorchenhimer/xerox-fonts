
SRC= 4word.go \
	 5word.go \
	 bdf.go \
	 character.go \
	 decode.go \
	 metadata.go \
	 font.go \
	 headers.go

CMDS= cmd/debug \
	  cmd/fnt2bdf

all: $(CMDS)

cmd/%: cmd/%.go $(SRC)
	go build -o $@ $<

clean:
	-rm decode
