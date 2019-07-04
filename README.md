# Redirector
Simple domain redirectors backed on go with sqlite.

There's also a simple web ui available

#### Run in docker:

    docker run -dv /local/data/path:/data \
        -p 1338:1338 \
        -e BASE_URL=redirector.mydomain.com \
        -e DB_PATH=/data \
        -e FRONT_PROXY=false \
        -e PORT=1338 \
        omropfryslan/redirector
