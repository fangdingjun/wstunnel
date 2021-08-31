wstunnel
=======

two features:

1. receive data through tcp and forward it through websocket connection

2. receive data through websocket and forward it to upstream tcp port


usage example
=====

### proxy ssh connection ###

on server side, listen for websocket connection and forward it to ssh port, example config

    proxy_config:
        - listen: ws://127.0.0.1:2222/p1
          remote: tcp://127.0.0.1:22

start server side proxy by `wstunnel -c server.yaml`

on client side, forward the connection through websocket, example config

    proxy_config:
        -  listen: tcp://127.0.0.1:9001
           remote: ws://127.0.0.1:2222/p1

start client side proxy by `wstunnel -c client.yaml`

now, you can connect to ssh server by `ssh -p 9001 127.0.0.1`, your ssh connection hide in websocket

the websocket can be proxied by nginx also



