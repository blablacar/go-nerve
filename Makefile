


app=nerve




############################################

check_defined = \
    $(foreach 1,$1,$(__check_defined))
__check_defined = \
    $(if $(value $1),, \
      $(error Undefined $1$(if $(value 2), ($(strip $2)))))


posturl=$(curl --data "{\"tag_name\": \"$1\",\"target_commitish\": \"master\",\"name\": \"$1\",\"body\": \"Release of version $1\",\"draft\": false,\"prerelease\": true}" https://api.github.com/repos/blablacar/go-nerve/releases?access_token=${access_token} | grep "\"upload_url\"" | sed -ne 's/.*\(http[^"]*\).*/\1/p')


all: clean utest build

build:
	godep go build -ldflags "-X main.BuildTime=`date -u '+%Y-%m-%d_%H:%M:%S_UTC'` -X main.Version=`cat VERSION.txt`-`git rev-parse HEAD`" -o nerve

clean:
	rm -f nerve

utest:
	godep go test ./...

install:
	cp nerve ${GOPATH}/bin/nerve


check_release:
	$(call check_defined, version, $(app) version)
	$(call check_defined, token, github access token)

release: check_release all

    check_git

	#git tag $version -a -m "Version $version"
	#git push --tags
	sleep 5
	curl -i -X POST -H "Content-Type: application/x-gzip" --data-binary "@${fullpath}" "${posturl%\{?name,label\}}?name=${filename}&label=${filename}&access_token=$(token)"
