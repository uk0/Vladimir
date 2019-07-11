#!/bin/sh
BASE=$(cd `dirname $0`; pwd)

mkdir -p /root/go


function get_librfkafka(){
yum -y install  openssl-devel cyrus-sasl cyrus-sasl-devel cyrus-sasl-lib  lz4-devel lz4 gcc-c++
    cd $BASE/librdkafka/ && ./configure && make && make install && cd $BASE
    export GOPATHVAL=`go env | grep GOPATH | awk -F '=' '{print$2}' | sed s/'"'//g`
}


echo "cd $BASE/../ && cp -r Octopoda $GOPATHVAL/src/github.com/uk0/"

function install_package(){
    rm -rf $GOPATHVAL/src/
    mkdir -p  $GOPATHVAL/src/
    mkdir -p $GOPATHVAL/src/github.com/uk0/
    cd $BASE/../ && cp -r Octopoda $GOPATHVAL/src/github.com/uk0/
    for i in `ls -la $BASE/import/ | grep ^d | awk 'NR>2{print$9}'`;
    do
         echo "$BASE/import/$i --> $GOPATHVAL/src/$i"
         cp -r $BASE/import/$i  $GOPATHVAL/src/
    done
}

function build_go(){

   cd $BASE/ && export PKG_CONFIG_PATH=/usr/local/lib/pkgconfig/ && ./build.sh
}

function chechVersion(){

source /etc/os-release
case $ID in
debian|ubuntu|devuan)
echo "debian|ubuntu|devuan"
    sudo apt-get install go
    ;;
centos|fedora|rhel)
echo "centos|fedora|rhel"
    yumdnf="yum"
    if test "$(echo "$VERSION_ID >= 22" | bc)" -ne 0; then
        yumdnf="dnf"
    fi
    sudo $yumdnf install -y go
    ;;
*)
    exit 1
    ;;
esac
}

function delete_all(){
    cd $BASE && rm -rf `ls -al  | grep -v 'release' | grep -v 'librdkafka' | awk '{print $9}'`
    cd $BASE/release && echo "执行 start.sh 启动采集" && ls
}

chechVersion

get_librfkafka

install_package

build_go

#delete_all