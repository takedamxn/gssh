GSSH=./gssh/gssh
GSCP=./gscp/gscp
PROGRAM=${GSSH} ${GSCP}

.PHONY: all
all:${PROGRAM}

${GSSH}:$(wildcard gssh/*.go common/*.go)
	go build -o $@ gssh/gssh

${GSCP}:$(wildcard gscp/*.go common/*.go)
	go build -o $@ gssh/gscp

.PHONY: clean
clean:
	rm -f ${PROGRAM}
