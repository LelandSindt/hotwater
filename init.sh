export GOPATH=/home/pi/go/
CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o main .
docker build . -t hotwater
sudo docker run --name hotwater --restart unless-stopped --privileged -p 8080:8080 -d hotwater 
