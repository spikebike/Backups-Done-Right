#!/bin/bash
rm etc/fs-meta.sql; go run walker.go --debug --new-db --threads=1 &> log; cat log | grep threads > r
rm etc/fs-meta.sql; go run walker.go --debug --new-db --threads=2 &> log; cat log | grep threads >> r
rm etc/fs-meta.sql; go run walker.go --debug --new-db --threads=4 &> log; cat log | grep threads >> r
