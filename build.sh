#!/bin/bash
set -e
start=`date +%s`
dir=$( dirname "$0" )

app=nerve
osarchi="$(go env GOHOSTOS)-$(go env GOHOSTARCH)"
[ -z "$1" ] || osarchi="$1"
[ ! -z ${version+x} ] || version="0"

[ -f ${GOPATH}/bin/godep ] || go get github.com/tools/godep
[ -f ${GOPATH}/bin/golint ] || go get github.com/tools/golint
[ -f /usr/bin/upx ] || (echo "upx is required to build" && exit 1)

${dir}/clean.sh

echo -e "\033[0;32mSave Dependencies\033[0m"
godep save ./${dir}/...

echo -e "\033[0;32mFormat\033[0m"
gofmt -w -s ${dir}/

echo -e "\033[0;32mLint\033[0m"
golint ${dir}

for e in `echo -e "$osarchi"`; do
    echo -e "\033[0;32mBuilding $e\033[0m"

    GOOS="${e%-*}" GOARCH="${e#*-}" \
    godep go build -ldflags "-X main.BuildTime=`date -u '+%Y-%m-%d_%H:%M:%S_UTC'` -X main.Version=${version}-`git rev-parse HEAD`" \
        -o $dir/dist/${app}-v${version}-${e}/${app}

    if [ "${e%-*}" != "darwin" ]; then
        echo -e "\033[0;32mCompressing ${e}\033[0m"
        upx ${dir}/dist/${app}-v${version}-${e}/${app}
    fi

    if [ "${e%-*}" == "windows" ]; then
        mv ${dir}/dist/${app}-v${version}-${e}/${app} ${dir}/dist/${app}-v${version}-${e}/${app}.exe
    fi
done

echo -e "\033[0;32mInstalling\033[0m"

cp ${dir}/dist/${app}-v${version}-$(go env GOHOSTOS)-$(go env GOHOSTARCH)/${app}* ${GOPATH}/bin/

echo "Duration : $((`date +%s`-start))s"
