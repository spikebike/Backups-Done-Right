#!/bin/bash
rm fs-meta.sql; go run walker.go --debug --new-db --threads=1 | grep threads > r
rm fs-meta.sql; go run walker.go --debug --new-db --threads=2 | grep threads >> r
rm fs-meta.sql; go run walker.go --debug --new-db --threads=4 | grep threads >> r
rm fs-meta.sql; go run walker.go --debug --new-db --threads=8 | grep threads >> r
