#!/bin/bash

VERSION=11.6.0

if [ ! -e debian.iso.wal ]
then
	rm -f debian.iso
fi
if [ ! -e SHA256SUMS ]
then
	curl -O https://cdimage.debian.org/debian-cd/${VERSION}/amd64/iso-cd/SHA256SUMS
fi

./bin/medusa-get debian.iso \
    http://debian.koyanet.lv/debian-cd/${VERSION}/amd64/iso-cd/debian-${VERSION}-amd64-netinst.iso \
    http://debian.anexia.at/debian-cd/${VERSION}/amd64/iso-cd/debian-${VERSION}-amd64-netinst.iso \
    http://ftp.crifo.org/debian-cd/${VERSION}/amd64/iso-cd/debian-${VERSION}-amd64-netinst.iso \
    http://debian.obspm.fr/debian-cd/${VERSION}/amd64/iso-cd/debian-${VERSION}-amd64-netinst.iso \
    https://ftp.cica.es/debian-cd/${VERSION}/amd64/iso-cd/debian-${VERSION}-amd64-netinst.iso \
    http://ftp.ps.pl/pub/Linux/debian-cd/${VERSION}/amd64/iso-cd/debian-${VERSION}-amd64-netinst.iso \
    http://debian.mirror.root.lu/debian-cd/${VERSION}/amd64/iso-cd/debian-${VERSION}-amd64-netinst.iso \
    http://ftp.lanet.kr/debian-cd/${VERSION}/amd64/iso-cd/debian-${VERSION}-amd64-netinst.iso \
    https://cdimage.debian.org/debian-cd/${VERSION}/amd64/iso-cd/debian-${VERSION}-amd64-netinst.iso

shasum -a 256 debian.iso
grep debian-${VERSION} SHA256SUMS
