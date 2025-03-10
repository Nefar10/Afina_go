#!/bin/bash
#Новый скрипт
  rm -rf Afina_go
  apt-get install git -y
  git clone https://github.com/Nefar10/Afina_go.git
  cp /go/Afina_go/init.sh /go/init.sh
  ln -s /go/Afina_go/downloads /go/downloads
  cd Afina_go
  go build -o main
  ./main -v