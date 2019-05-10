#!/bin/bash

DEBEMAIL="aozarkar@equinix.com"
DEBFULLNAME="Anand Ozarkar"
export DEBEMAIL DEBFULLNAME

dh_make -s --createorig -y -p smartkey-kmsplugin_1.0

echo "smartkey-kms /usr/bin
conf/smartkey-grpc.service /lib/systemd/system/
conf/smartkey-grpc.conf /etc/smartkey/
conf/smartkey.yaml /etc/smartkey/" > debian/install

debuild -us -uc
