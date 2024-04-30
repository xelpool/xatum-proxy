
rm build -r

mkdir build
cp LICENSE.txt ./build/
cd build

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -trimpath -ldflags="-s -w" ..

tar -czf xatum-proxy-linux.tar.gz xatum-proxy LICENSE.txt

GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w" ..

zip xatum-proxy-windows.zip xatum-proxy.exe LICENSE.txt
