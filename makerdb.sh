#!/bin/bash

docker run --rm --name my-redis -d redis

docker exec -it my-redis redis-cli set fo b
docker exec -it my-redis redis-cli set foo bar
docker exec -it my-redis redis-cli set adsf dd EX 10000
docker exec -it my-redis redis-cli set a2sf dd PX 234332
docker exec -it my-redis redis-cli set ssd dd PX 53422
docker exec -it my-redis redis-cli save

# docker exec -it my-redis redis-cli set empty ""
# docker exec -it my-redis redis-cli lpush mylist a b c
# docker exec -it my-redis redis-cli sadd myset x y z
# docker exec -it my-redis redis-cli zadd myzset 1 one 2 two -1 neg 2.5 float
# docker exec -it my-redis redis-cli hset myhash field1 value1 field2 value2
# docker exec -it my-redis redis-cli xadd mystream * field1 value1
# docker exec -it my-redis redis-cli set "sp ace" "sp ace value"
# docker exec -it my-redis redis-cli set "uni键" "值"
# docker exec -it my-redis redis-cli set expireme foo EX 1
# docker exec -it my-redis redis-cli pfadd myhll a b c
# docker exec -it my-redis redis-cli geoadd mygeo 13.361389 38.115556 "Palermo"
# docker exec -it my-redis redis-cli setbit mybitmap 100 1

rm ./dump.rdb
docker cp my-redis:/data/dump.rdb ./dump.rdb
docker stop my-redis
