# go-directequality-checker

go-directequality-checker is a small tool that performs static analysis of a Go package to identify places where direct equality comparisons of fields are done when it isn't appropriate to do so. This is particularly useful if you're wanting to ensure that certain fields are only compared using constant-time comparison for security reasons.

## Installation
Installing it is as simple as running

```
go get github.com/1password/go-directequality-checker
```

## Usage
In order to use this tool, you'll first need to annotate a struct field that you want to designate as not being allowed to do direct equality comparisons with a tag: `security:"nodirectequality"`

Example:
```
type User struct {
    VerificationToken string `db:"verification_token" security:"nodirectequality"`
}
```

Once you've annotated your field, you can run the tool via

```
go-directequality-checker path/to/go/package
```

If the tool finds any direct equality comparisons, it will output information about what it has found and recommend using a constant time comparison function (i.e. crypto/subtle's `ConstantTimeCompare`):

```
[SECURITY] Found raw comparison of field 'VerificationToken'. Use constant time comparison function.
/Users/rfillion/go/src/go.1password.io/b5/server/src/logic/action/transfer.go:106
user.VerificationToken == token {
```