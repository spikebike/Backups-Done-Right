# Backups-Done-Right
P2P backup program

## Table of Contents
- [Client](#Client)
- [Server](#Server)
- [DHT](#DHT)

##Client

The client walks the configured directories tracking any deletions, updates, or additions.  Any changes are compressed, encrypted, and queued for upload.  Deduplication happens in the process of uploading.

##Server

The server receives SHA256 checksums from the client and accepts uploads for blobs it hasn't seen before.  Then applies a reed-solomon error correction code to generate chunks, which are traded with other peers.  The server works to ensure the specified replication is maintained and periodically challenges other peers to ensure they are storing the agreed upon chunks.

