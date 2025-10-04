#!/bin/bash

rm mem-db.tar
docker save mem-db:latest > mem-db.tar
scp ./mem-db.tar ofl:/home/ofl/MEM-db
