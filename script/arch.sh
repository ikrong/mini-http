rm -rf ./dist

# osarchs=$(go tool dist list | grep -E 'windows|linux|darwin')
osarchs="darwin/amd64
darwin/arm64
linux/amd64
linux/arm64
linux/arm/v7
linux/arm/v6
linux/arm/v5
linux/386
windows/386
windows/amd64
windows/arm/v7"

currentOsName=$(uname -s | tr A-Z a-z)

currentOsArch=$(dpkg --print-architecture)

currentUpxVersion=$(wget -qO- -t1 -T2 "https://api.github.com/repos/upx/upx/releases/latest" | grep "tag_name" | head -n 1 | awk -F ":" '{print $2}' | sed 's/\"//g;s/,//g;s/ //g' | tr -d a-zA-Z)

upxFileName="upx-${currentUpxVersion}-${currentOsArch}_${currentOsName}"

upxDownloadUrl="https://github.com/upx/upx/releases/download/v${currentUpxVersion}/${upxFileName}.tar.xz"

echo $upxDownloadUrl

if [ ! -f "./upx/$upxFileName/upx" ]; then
    mkdir -p ./upx
    wget $upxDownloadUrl -O ./upx/upx.tar.xz
    tar -xvf ./upx/upx.tar.xz -C ./upx
fi

for osarch in $osarchs
do
    info=(${osarch//// })
    os=${info[0]}
    arch=${info[1]}
    ver=${info[2]}
    echo "building $os $arch $ver"
    if [ "$arch"="arm" -a -n "$ver" ];then
        GOOS=$os GOARCH=$arch GOARM=$(echo $ver | tr -d a-zA-Z) CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/$osarch/serve main.go
        ./upx/$upxFileName/upx --best dist/$osarch/serve
        if [ "$os"="windows" ];then
            zip -j dist/${os}_${arch}_${ver}_serve.zip dist/$osarch/serve
        else
            tar -czvf dist/${os}_${arch}_${ver}_serve.tar.gz -C dist/$osarch dist/$osarch/serve
        fi
    else
        GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/$osarch/serve main.go
        ./upx/$upxFileName/upx --best dist/$osarch/serve
        if [ "$os"="windows" ];then
            zip -j dist/${os}_${arch}_serve.zip dist/$osarch/serve
        else
            tar -czvf dist/${os}_${arch}_serve.tar.gz -C dist/$osarch dist/$osarch/serve
        fi
    fi
    echo "$os $arch $ver build finished"
done

