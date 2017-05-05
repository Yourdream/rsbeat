FROM busybox

COPY docker-entrypoint.sh /docker-entrypoint.sh
COPY rsbeat-linux-amd64 /rsbeat
COPY rsbeat.yml /rsbeat.yml
COPY rsbeat.template.json /rsbeat.template.json
COPY rsbeat.template-es2x.json /rsbeat.template-es2x.json

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["/rsbeat", "-e", "-d", "*", "-c", "/rsbeat.yml"]
