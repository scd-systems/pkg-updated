UNAME:=$(shell uname)

PRJNAME=pkg-updated
TEXSRC_FILE=pkg-updated.nw
TEX_FILE=pkg-updated.tex
PROGRAM=pkg-updated.go
CONFIG=pkg-updated.conf
TODO=TODO
MAN=pkg-updated-manpage
major_version:=$(shell grep 'MAJOR_VERSION' $(TEXSRC_FILE) | cut -d"=" -f2 | head -n 1)
minor_version:=$(shell grep 'MINOR_VERSION' $(TEXSRC_FILE) | cut -d"=" -f2 | head -n 1)
patch_version:=$(shell grep 'PATCH_VERSION' $(TEXSRC_FILE) | cut -d"=" -f2 | head -n 1)
date:=$(shell date)
pwd:=$(shell pwd)

all: extract_debug run

extract: 
	@notangle -R${PROGRAM} $(TEXSRC_FILE) > ${PROGRAM}
	@notangle -R${CONFIG} $(TEXSRC_FILE) > ${CONFIG}
	@notangle -R${TODO} $(TEXSRC_FILE) > ${TODO}
	@notangle -R${MAN} $(TEXSRC_FILE) > ${MAN}
	@go fmt ${PROGRAM}
	@chmod 755 ${PROGRAM}

extract_debug: 
	@notangle -L'// %L "%F"%N' -R${PROGRAM} $(TEXSRC_FILE) > ${PROGRAM}
	@notangle -R${CONFIG} $(TEXSRC_FILE) > ${CONFIG}
	@notangle -R${TODO} $(TEXSRC_FILE) > ${TODO}
	@go fmt ${PROGRAM}
	@chmod 755 ${PROGRAM}

parse:
	sed -ie "s/SCRIPTBUILDDATE/`date`/" ${PROGRAM}

test:
	go tool vet -all ${PROGRAM}
	go tool vet -shadow ${PROGRAM}

pdf:
	@noweave -autodefs c -index -delay ${TEXSRC_FILE} > content.tex
	@cp titlepage.tpl titlepage.tex
	@perl -pi -e 's/__VERSION__/${major_version}\.${minor_version}\.${patch_version}/g' titlepage.tex
	@perl -pi -e 's/__DATE__/${date}/g' titlepage.tex
	@pdflatex ${TEX_FILE} >/dev/null
	@pdflatex ${TEX_FILE} >/dev/null
	@pdflatex ${TEX_FILE} >/dev/null
	@echo "PDF File: ${PRJNAME}.pdf are successfully created"
man:
	# Make manpage tasks
	
clean:
	rm ./pkg-updated.go ./*.aux ./*.log ./*.out ./*.pdf ./*.toc ./titlepage.tex

run: all
	sleep 1
