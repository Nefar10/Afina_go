#!/bin/bash
  rm -rf Afina_go
  apt-get install git -y
  git clone https://github.com/Nefar10/Afina_go.git
  cp /go/Afina_go/init.sh /go/init.sh
  cd Afina_go
  go build -o main
  ./main -v