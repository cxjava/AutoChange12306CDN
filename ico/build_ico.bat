@Echo set http_proxy=http://127.0.0.1:1080
@Echo set https_proxy=http://127.0.0.1:1080
@Echo go get -u github.com/akavel/rsrc
rsrc -manifest="ico.manifest" -ico="1.ico" -o ico.syso
pause