
SRC= 4word.go \
	 5word.go \
	 character.go \
	 decode.go \
	 metadata.go \
	 font.go \
	 headers.go

CMDS= \
	  cmd/debug

all: $(CMDS)

cmd/debug: cmd/debug.go $(SRC)
	go build -o $@ $<

#decode: $(SRC)
#	go build -o $@

clean:
	-rm decode
