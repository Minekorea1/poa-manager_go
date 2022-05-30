#!/bin/bash

first=true
for file in **/*; do
    filename=${file##*/}
    filename=${filename%.*}

    if $first; then
        fyne bundle -package res -name ${filename^} $file > bundle.go
        first=false
    else
        fyne bundle -append -name ${filename^} $file >> bundle.go
    fi
done

# fyne bundle -package res -name NanumBarunGothicTTF fonts/NanumBarunGothic.ttf > bundle.go
# fyne bundle -append      -name IconMain            image/icon_main.png >> bundle.go