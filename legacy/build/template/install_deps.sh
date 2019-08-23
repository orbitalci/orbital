#!/usr/bin/env sh
command -v source
if [ $? != 0 ]; then
    ID=$(cat /etc/os-release | grep -w ID | cut -d "=" -f2)
else
    source /etc/os-release
fi

get_installerupdater ()
{
  if [ $ID = "alpine" ]; then
    echo "updating pkg manager apk"
    apk update
    installer="apk add"
  elif [ $ID = "debian" -o $ID = "ubuntu" ]; then
    echo "updating pkg manager apt"
    apt update
    installer="apt install -y"
  else 
    echo "Ocelot does not yet support this distro. Please contact prod-eng."
    exit 1
  fi
} 

get_installerupdater

download_dep() 
{   
echo "$1"
command -v "$1"
if [ $? != 0 ]; then
    echo "dependency $1 not found, attempting to install"
    if [ $ID = "alpine" ]; then
        apk add $1
    elif [ $ID = "debian" -o $ID = "ubuntu" ]; then
        apt install -y $1
    fi
else
    echo "found dependency $1"
fi

}

dependencies="openssl bash zip curl wget git python base64"
for dep in $dependencies; do 
    download_dep $dep
done