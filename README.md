### mitum-did

*mitum-did* is the mitum model for operating Protocon DID, based on
[*mitum*](https://github.com/spikeekips/mitum) and [*mitum-currency*](https://github.com/spikeekips/mitum-currency).

#### Features,

* account: account address and keypair is not same.
* documentData: actual data stored in document.
* simple transaction: create document.
* supports multiple keypairs: *btc*, *ethereum*, *stellar* keypairs.
* *mongodb*: as mitum does, *mongodb* is the primary storage.

#### Installation

> NOTE: at this time, *mitum* and *mitum-did* is actively developed, so
before building mitum-blocksign, you will be better with building the latest
mitum source.
> `$ git clone https://github.com/ProtoconNet/mitum-did.git`
>
> and then, add `replace github.com/spikeekips/mitum => <your mitum source directory>` to `go.mod` of *mitum-did*.

Build it from source
```sh
$ cd mitum-did
$ go build -ldflags="-X 'main.Version=v0.0.1'" -o ./md ./main.go
```

#### Run

At the first time, you can simply start node with example configuration.

> To start, you need to run *mongodb* on localhost(port, 27017).

```
$ ./md node init ./standalone.yml
$ ./md node run ./standalone.yml
```

> Please check `$ ./md --help` for detailed usage.

#### Test

```sh
$ go clean -testcache; time go test -race -tags 'test' -v -timeout 20m ./... -run .
```
