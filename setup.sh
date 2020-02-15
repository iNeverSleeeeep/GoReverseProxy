rm -f /usr/bin/gorproxy
rm -f /etc/init.d/gorproxy
rm -f /etc/gorproxy.yaml
go build -o gorproxy main.go
ln gorproxy /usr/bin/gorproxy
touch /etc/init.d/gorproxy
echo "gorproxy &" >> /etc/init.d/gorproxy
chmod 777 /etc/init.d/gorproxy
cp -f ./gorproxy.yaml /etc/gorproxy.yaml
service gorproxy start