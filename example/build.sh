#!/bin/zsh

export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0

home=$PWD
root="$home/cmd"

cd $root

for i in $(ls -d **/)
do
    if [[ -f "$i/main.go" ]]
    then
        cd $i

        out="$home/bin/${i%/}"
        build=$(go build -v -o ${out} 2>&1)
        
        echo $out
        bootstrap="$home/bin/bootstrap"
        cat $out > $bootstrap
        zip -jD "$out".zip $bootstrap
        
        cd $root
        exit
    fi
done

cd "$home"
