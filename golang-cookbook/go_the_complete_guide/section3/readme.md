#
## Run
```
go run main.go deck.go
```

## Test
```
go test
```

## Error
### cannot find main module
원인: go test 명령어 실행 후 아래와 같은 에러 반환      
```shell
go test
#go: cannot find main module, but found .git/config in /Users/hangeulbae/Desktop/golang-cookbook
#        to create a module there, run:
#        cd .. && go mod init
```
해결      
```shell
# go env 명령어 실행 후 GO111MODULE=""으로 되어 있으면 아래와 같이 명령어를 실행하여 해결
go env
go env -w GO111MODULE=auto
```