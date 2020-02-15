rm -f /usr/bin/gorproxy
rm -f /etc/init.d/gorproxy
rm -f /etc/gorproxy.yaml
go build -o gorproxy main.go
ln gorproxy /usr/bin/gorproxy
cp -f gorproxy /etc/init.d/gorproxy
cp -f ./gorproxy.yaml /etc/gorproxy.yaml