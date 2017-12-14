PROGRAM = gssh gscp
all:${PROGRAM}

gssh:gssh.go common/config.go
	go build $@.go

gscp:gscp.go common/config.go
	go build $@.go

clean:
	rm -f ${PROGRAM}
