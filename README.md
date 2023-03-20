# Medusa

Fetch one big file from multiple URL.

## Try it

```bash
make
./bin/medusa-get url…
```

Don't forget to shasum the downloaded file, if mirrors are out of sync.

The `demo.sh` script download a Debian image and validate it.

## Explainations

It's a poor man peer to peer download, working with plain old HTTP.
It's great for lower bandwidth of mirrors.

S3 and its clone love this pattern, download can be shared between replicas.

```
https://cdimage.debian.org/debian-cd/${VERSION}/amd64/iso-cd/debian-${VERSION}-amd64-netinst.iso
```
Firts URL must be a complete URL. Output file is guessed from this URL.
```
http://ftp.ps.pl/pub/Linux/debian-cd
http://debian.mirror.root.lu/debian-cd
…
```
All others URL are mirror. The last element of a mirror must be the starting element of the first URL.

All URL are tested with a HEAD, and size file is checked.
Some server doesn't like multiple connection, and answer 429.

The file is split in 1Mb chunks.
The `Range` header is used to pick the right chunk.
Connection is done with a timeout.
Downloading a chunk has an other timeout.
Failed download destroy the downloader.

A WAL (Write Ahead Log) is used, and you can stop and restart a downnload.
