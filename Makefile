GOCC=go
USER=`echo $${USER}`

# To Compile the linux version using docker simply invoke the makefile like this:
#
# make GOCC="docker run --rm -t -v ${GOPATH}:/go hbouvier/go-lang:1.5"
PROJECTNAME=monitor

all: get-deps fmt darwin linux arm build coverage

clean:
	rm -rf coverage.out \
				 ${GOPATH}/pkg/{linux_amd64,darwin_amd64,linux_arm}/github.com/hbouvier/${PROJECTNAME} \
				 ${GOPATH}/bin/{linux_amd64,darwin_amd64,linux_arm}/${PROJECTNAME} \
				 release

build: fmt test
	${GOCC} install github.com/hbouvier/${PROJECTNAME}

fmt:
	${GOCC} fmt github.com/hbouvier/${PROJECTNAME}

test:
	# ${GOCC} test -v -cpu 4 -count 1 -coverprofile=coverage.out github.com/hbouvier/${PROJECTNAME}

coverage:
	# ${GOCC} tool cover -html=coverage.out

get-deps:
	${GOCC} get github.com/c9s/goprocinfo/linux github.com/hashicorp/logutils github.com/hbouvier/httpclient

linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 ${GOCC} install github.com/hbouvier/${PROJECTNAME}
	@if [[ $(shell uname | tr '[:upper:]' '[:lower:]') == $@ ]] ; then mkdir -p ${GOPATH}/bin/$@_amd64 && mv ${GOPATH}/bin/${PROJECTNAME} ${GOPATH}/bin/$@_amd64/ ; fi

darwin:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 ${GOCC} install github.com/hbouvier/${PROJECTNAME}
	@if [[ $(shell uname | tr '[:upper:]' '[:lower:]') == $@ ]] ; then mkdir -p ${GOPATH}/bin/$@_amd64 && mv ${GOPATH}/bin/${PROJECTNAME} ${GOPATH}/bin/$@_amd64/ ; fi

arm:
	GOOS=linux GOARCH=arm CGO_ENABLED=0 ${GOCC} install github.com/hbouvier/${PROJECTNAME}
	@if [[ $(shell uname | tr '[:upper:]' '[:lower:]') == $@ ]] ; then mkdir -p ${GOPATH}/bin/$@_amd64 && mv ${GOPATH}/bin/${PROJECTNAME} ${GOPATH}/bin/$@_amd64/ ; fi

release: linux darwin arm
	@mkdir -p release/bin/{linux_amd64,darwin_amd64,linux_arm}
	for i in linux_amd64 darwin_amd64 linux_arm; do cp ${GOPATH}/bin/$${i}/${PROJECTNAME} release/bin/$${i}/ ; done
	COPYFILE_DISABLE=1 tar cvzf release/${PROJECTNAME}.v`cat VERSION`.tgz release/bin
	zip -r release/${PROJECTNAME}.v`cat VERSION`.zip release/bin

container: release
	docker build -t ${USER}/go-monitor .