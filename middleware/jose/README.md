# Jose Middleware for Iris web framework

[![GoDoc](https://godoc.org/github.com/kardianos/govendor?status.svg)](https://godoc.org/github.com/Heirko/iris-contrib/middleware/jose)

Golang jose repository

## Middleware information

This middleware has been converted to work with [iris](https://github.com/kataras/iris)


## Description

Middleware for Javascript Object Signing and Encryption.
The implementation follows the
[JSON Web Encryption](http://dx.doi.org/10.17487/RFC7516)
standard (RFC 7516) and
[JSON Web Signature](http://dx.doi.org/10.17487/RFC7515)
standard (RFC 7515).

Underlying library is [go-jose from Square](https://github.com/square/go-jose)


## Install

```sh
$ go get -u github.com/Heirko/iris-contrib/middleware/jose
```

## How to use

Read the jwt middleware section [here](https://kataras.gitbooks.io/iris/content/jwt.html)
