
```
go generate .
go-fuzz-build github.com/lugu/audit/fuzz
go-fuzz -bin=./fuzz-fuzz.zip -workdir .
```
