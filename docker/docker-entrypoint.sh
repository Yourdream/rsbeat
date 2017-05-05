#!/bin/sh
set -e

if [ ! $ES_URL ];then
    ES_URL="elasticsearch:9200"
fi

if [ ! $REDIS_LIST ];then
    REDIS_LIST='"10.0.0.40:6379"'
fi

if [ ! $REDIS_SLOWER_THAN ];then
    REDIS_SLOWER_THAN=200
fi

if [ ! $PERIOD ];then
    PERIOD="1s"
fi

# Render config file
cat rsbeat.yml | sed "s|ES_URL|${ES_URL}|g" | sed "s|REDIS_LIST|${REDIS_LIST}|g" | sed "s|REDIS_SLOWER_THAN|${REDIS_SLOWER_THAN}|g" | sed "s|PERIOD|${PERIOD}|g"  > rsbeat.yml.tmp
cat rsbeat.yml.tmp > rsbeat.yml
rm rsbeat.yml.tmp

exec "$@"