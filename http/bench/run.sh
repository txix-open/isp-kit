bombardier -c 128 -t 1s -d 15s --fasthttp -m POST -f body.json http://localhost:8000/post
bombardier -c 128 -t 1s -d 15s --fasthttp -m POST -f body.json http://localhost:8001/post
bombardier -c 128 -t 1s -d 15s --fasthttp -m POST -f body.json http://localhost:8002/post
