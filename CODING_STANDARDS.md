# TilePix coding standards
To ensure the codebase remains consistant, these standards are a requirement.  These standards are a modified version
of those from the Go blog.

These standards are subject to change.

## Comment sentences
See [effective Go commentary](https://golang.org/doc/effective_go.html#commentary). Comments documenting declarations
should be full sentences, even if that seems a little redundant. This approach makes them format well when extracted
into godoc documentation. Comments should begin with the name of the thing being described and end in a period:
```go
// Request represents a request to run a command.
type Request struct { ...

// Encode writes the JSON encoding of req to w.
func Encode(w io.Writer, req *Request) { ...
```

All exported variables, constants, interfaces and functions require documentation.

## Declaring empty slices
When declaring an empty slice, prefer
```go
var t []string
```
over
```go
t := []string{}
```

The former declares a nil slice value, while the latter is non-nil but zero-length. They are functionally
equivalent -- their `len` and `cap` are both zero -- but the nil slice is the preferred style.

Note that there are limited circumstances where a non-nil but zero-length slice is preferred, such as when encoding JSON
objects (a `nil` slice encodes to `null`, while `[]string{}` encodes to the JSON array `[]`).

When designing interfaces, avoid making a distinction between a nil slice and a non-nil, zero-length slice, as this can
lead to subtle programming errors.

## Crypto rand
Do not use package math/rand to generate keys, even throwaway ones. Unseeded, the generator is completely predictable.
Seeded with time.Nanoseconds(), there are just a few bits of entropy. Instead, use crypto/rand's Reader, and if you need
text, print to hexadecimal or base64:

```go
import (
   "crypto/rand"
   // "encoding/base64"
   // "encoding/hex"
   "fmt"
)

func Key() string {
   buf := make([]byte, 16)
   _, err := rand.Read(buf)
   if err != nil {
       panic(err)  // out of randomness, should never happen
   }
   return fmt.Sprintf("%x", buf)
   // or hex.EncodeToString(buf)
   // or base64.StdEncoding.EncodeToString(buf)
}
```

## Don't panic
See [effective Go errors](https://golang.org/doc/effective_go.html#errors). Don't use panic for normal error handling.
Use error and multiple return values.

## Error checking
No function which potentially returns an error should be unassigned.  Where possible check the return value and
propogate as soon as possible.

### Blank assignments
Occasionally we have to call a function which returns a potential error, after we have already caught an error.  It is
only under these circumstances we should assign the return value to `_`.  For example, when work with transations:
```go
tx, err := newTransaction()
if err != nil {
	return err
}

if err := tx.PerformTask(1); err != nil {
	_ = tx.Rollback()
	return err
}

if err := tx.PerformTask(2); err != nil {
	_ = tx.Rollback()
	return err
}

return tx.Commit()
```

### Deferred checks
When calling a function which returns an error using defer, also call the function (if possible) explicitally, checking
the error. Example:
```go
f, err := os.Open("test.txt")
if err != nil {
	return err
}
// Defer close, because we *must* close the file
defer f.Close()

// ... more code which may or may not return from the function early with errors

// As soon as we don't need to use `f` again
if err := f.Close(); err != nil {
	return err
}

// ...

return nil
```

### Naming
Any variables holding error messages should be prefixed with `Err`.  Example:
```go
var (
	ErrFooUninitiatlise = errors.New("foo must first be initialised")
)
```

Any custom error types should be suffixed with `Error`.  Example:
```go
// ParseError is type of error returned when there's a parsing problem.
type ParseError struct {
  Line, Col int
}

func foo() {
    res, err := somepkgAction()
    if err != nil {
        if err == somepkg.ErrBadAction {
        	// ...
        }
        if pe, ok := err.(*ParseError); ok {
             line, col := pe.Line, pe.Col
             // ...
        }
    }
}
```

It should always be safe to call `.Close()` according to the `Closer` interface.  If writing a struct which implements
`io.Closer`, ensure it is safe to call `Close` multiple times.

## Error strings
Error strings should not be capitalized (unless beginning with proper nouns or acronyms) or end with punctuation, since
they are usually printed following other context. That is, use `errors.New("something bad")` not
`errors.New("Something bad")`, so that `log.Printf("Reading %s: %v", filename, err)` formats without a spurious capital
letter mid-message. This does not apply to logging, which is implicitly line-oriented and not combined inside other
messages.

When you require to insert data into the error message use `fmt.Errorf()`, however prefer `errors.New()` and use logging
to including the extra data.

## Examples
When adding a new package, include examples of intended usage: a runnable Example, or a simple test demonstrating a
complete call sequence.

Read more about [testable Example() functions](https://blog.golang.org/examples).

## Goimports
Run [goimports](https://godoc.org/golang.org/x/tools/cmd/goimports) on your code to automatically fix the majority of
mechanical style issues.

## Imports
Avoid renaming imports except to avoid a name collision; good package names should not require renaming. In the event of
collision, prefer to rename the most local or project-specific import.

Imports are organized in groups, with blank lines between them. The standard library packages are always in the first
group.
```go
package main

import (
    "fmt"   
    "hash/adler32"
    "os"

    "appengine/foo"
    "appengine/user"

    "github.com/foo/bar"
    
    "rsc.io/goversion/version"
)
```

## Import blank
Packages that are imported only for their side effects (using the syntax `import _ "pkg"`) should only be imported with
a comment stating why the blank import is required.

## In-band errors
In C and similar languages, it's common for functions to return values like -1 or null to signal errors or missing
results:
```go
// Lookup returns the value for key or "" if there is no mapping for key.
func Lookup(key string) string

// Failing to check a for an in-band error value can lead to bugs:
Parse(Lookup(key))  // returns "parse failure for value" instead of "no value for key"
```

Go's support for multiple return values provides a better solution. Instead of requiring clients to check for an in-band
error value, a function should return an additional value to indicate whether its other return values are valid. This
return value may be an error, or a boolean when no explanation is needed. It should be the final return value.
```go
// Lookup returns the value for key or ok=false if there is no mapping for key.
func Lookup(key string) (value string, ok bool)
```

This prevents the caller from using the result incorrectly:
```go
Parse(Lookup(key))  // compile-time error
```

And encourages more robust and readable code:
```go
value, ok := Lookup(key)
if !ok  {
   return fmt.Errorf("no value for %q", key)
}
return Parse(value)
```

Return values like nil, "", 0, and -1 are fine when they are valid results for a function, that is, when the caller need
not handle them differently from other values.

Some standard library functions, like those in package "strings", return in-band error values. This greatly simplifies
string-manipulation code at the cost of requiring more diligence from the programmer. In general, Go code should return
additional values for errors.

## Indent error flow
Try to keep the normal code path at a minimal indentation, and indent the error handling, dealing with it first. This
improves the readability of the code by permitting visually scanning the normal path quickly. For instance, don't write:
```go
if err != nil {
    // error handling
} else {
    // normal code
}
```

Instead, write:
```go
if err != nil {
    // error handling
    return // or continue, etc.
}
// normal code
```

If the if statement has an initialization statement, such as:
```go
if x, err := f(); err != nil {
    // error handling
    return
} else {
    // use x
}
```

then this may require moving the short variable declaration to its own line:
```go
x, err := f()
if err != nil {
    // error handling
    return
}
// use x
```

## Initialisms
Words in names that are initialisms or acronyms (e.g. "URL" or "NATO") have a consistent case. For example, "URL" should
appear as "URL" or "url" (as in "urlPony", or "URLPony"), never as "Url". As an example: ServeHTTP not ServeHttp. For
identifiers with multiple initialized "words", use for example "xmlHTTPRequest" or "XMLHTTPRequest".

Code generated by tools such as the protocol buffer compiler is exempt from this rule. Human-written code is held to a
higher standard than machine-written code.

## Language
All comments and documentation must be in English (UK).  Spellings and grammatical conventions from other languages or
dialects will not be accepted.

These include:
 - The spelling of "colour", "behaivour", "capitilisation" etc.
 - Titles having the same capitilisation rules of sentences. 

## Line length
There is no rigid line length limit in Go code, but avoid uncomfortably long lines. Similarly, don't add line breaks to
keep lines short when they are more readable long--for example, if they are repetitive.

In other words, break lines because of the semantics of what you're writing (as a general rule) and not because of the
length of the line. If you find that this produces lines that are too long, then change the names or the semantics and
you'll probably get a good result.

This is, actually, exactly the same advice about how long a function should be. There's no rule "never have a function
more than N lines long", but there is definitely such a thing as too long of a function, and of too stuttery tiny
functions, and the solution is to change where the function boundaries are, not to start counting lines.

For comments lines must be limited to 120 characters.

## Logging
Always use the [logrus](https://github.com/sirupsen/logrus).

Prefer more `Debug`/`Trace` log calls than fewer.  It saves huge amounts of time when future debugging needs to be done.

Use the correct level of log call; do not log an error using a `Info` call, for example.

Log lines should have the structure `FunctionName: message`, and for methods: `StructName.MethodName: message`.
Example:
```go
func GetFoo() {
	log.Info("GetFoo: getting foo")
}

func (c client) Enable() {
	log.Info("client.Enable: enabling client")
}
```

The message we log should **only** ever be a static string, use `fields` to include dynamic data:
```go
// Bad
log.Info(fmt.Sprintf("Function(): event: %s", event.String()))

// Good
log.WithField("Event", event).Info("Function(): An event occured")
```

Use `WithField` when only one field is required, if required use `WithFields(log.Fields{})` to log more data.  Use
`WithError` for errors.
```go
func Do(i, j int, foo string) {
	// ...
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"i": i, "j": j, "foo": foo}).Error("Do: ...")
	}
	// ...
}
```

## Mixed caps
See [effective Go mixed-caps](https://golang.org/doc/effective_go.html#mixed-caps). This applies even when it breaks
conventions in other languages. For example an unexported constant is `maxLength` not `MaxLength` or `MAX_LENGTH`.

## Named Result Parameters
Consider what it will look like in godoc. Named result parameters like:
```go
func (n *Node) Parent1() (node *Node)
func (n *Node) Parent2() (node *Node, err error)
```

will stutter in godoc; better to use:
```go
func (n *Node) Parent1() *Node
func (n *Node) Parent2() (*Node, error)
```

On the other hand, if a function returns two or three parameters of the same type, or if the meaning of a result isn't
clear from context, adding names may be useful in some contexts. Don't name result parameters just to avoid declaring a
var inside the function; that trades off a minor implementation brevity at the cost of unnecessary API verbosity.
```go
func (f *Foo) Location() (float64, float64, error)
```

is less clear than:
```go
// Location returns f's latitude and longitude.
// Negative values mean south and west, respectively.
func (f *Foo) Location() (lat, long float64, err error)
```

Naked returns are never okay.

Finally, in some cases you need to name a result parameter in order to change it in a deferred closure. That is always
OK.

## Package comments
Package comments, like all comments to be presented by godoc, must appear adjacent to the package clause, with no blank
line.
```go
// Package math provides basic constants and mathematical functions.
package math
```

```go
/*
Package template implements data-driven templates for generating textual
output such as HTML.
....
*/
package template
```

Note that starting the sentence with a lower-case word is not among the acceptable options for package comments, as
these are publicly-visible and should be written in proper English, including capitalizing the first word of the
sentence. When the binary name is the first word, capitalizing it is required even though it does not strictly match the
spelling of the command-line invocation.

See [effective Go commentary](https://golang.org/doc/effective_go.html#commentary) for more information about commentary
conventions.

## Package names
All references to names in your package will be done using the package name, so you can omit that name from the
identifiers. For example, if you are in package `chubby`, you don't need type `ChubbyFile`, which clients will write as
`chubby.ChubbyFile`. Instead, name the type `File`, which clients will write as `chubby.File`. Avoid meaningless
package names like util, common, misc, api, types, and interfaces. See
[effective Go package-names](http://golang.org/doc/effective_go.html#package-names) for more.
   
## Pass values
Don't pass pointers as function arguments just to save a few bytes. If a function refers to its argument `x` only as
`*x` throughout, then the argument shouldn't be a pointer. Common instances of this include passing a pointer to a
string (`*string`) or a pointer to an interface value (`*io.Reader`). In both cases the value itself is a fixed size and
can be passed directly. This advice does not apply to large structs, or even small structs that might grow.

Do not pass pointers to interfaces, ever.

## Receiver names
The name of a method's receiver should be a reflection of its identity; often a one or two letter abbreviation of its
type suffices (such as `c` or `cl` for `Client`). Don't use generic names such as `me`, `this` or `self`, identifiers
typical of object-oriented languages that gives the method a special meaning. In Go, the receiver of a method is just
another parameter and therefore, should be named accordingly. The name need not be as descriptive as that of a method
argument, as its role is obvious and serves no documentary purpose. It can be very short as it will appear on almost
every line of every method of the type; familiarity admits brevity. Be consistent, too: if you call the receiver `c` in
one method, don't call it `cl` in another.

## Receiver type
Choosing whether to use a value or pointer receiver on methods can be difficult. If in doubt, use a pointer, but there
are times when a value receiver makes sense, usually for reasons of efficiency, such as for small unchanging structs or
values of basic type. Some useful guidelines:

 - If the receiver is a `map`, `func` or `chan`, don't use a pointer to them. If the receiver is a slice and the method
 doesn't reslice or reallocate the slice, don't use a pointer to it.
 - If the method needs to mutate the receiver, the receiver must be a pointer.
 - If the receiver is a struct that contains a sync.Mutex or similar synchronizing field, the receiver must be a pointer
 to avoid copying.
 - If the receiver is a large struct or array, a pointer receiver is more efficient. How large is large? Assume it's
 equivalent to passing all its elements as arguments to the method. If that feels too large, it's also too large for the
 receiver.
 - Can function or methods, either concurrently or when called from this method, be mutating the receiver? A value type
 creates a copy of the receiver when the method is invoked, so outside updates will not be applied to this receiver. If
 changes must be visible in the original receiver, the receiver must be a pointer.
 - If the receiver is a struct, array or slice and any of its elements is a pointer to something that might be mutating,
 prefer a pointer receiver, as it will make the intention more clear to the reader.
 - If the receiver is a small array or struct that is naturally a value type (for instance, something like the time.Time
 type), with no mutable fields and no pointers, or is just a simple basic type such as int or string, a value receiver
 makes sense. A value receiver can reduce the amount of garbage that can be generated; if a value is passed to a value
 method, an on-stack copy can be used instead of allocating on the heap. (The compiler tries to be smart about avoiding
 this allocation, but it can't always succeed.) Don't choose a value receiver type for this reason without profiling
 first.
 
## Structs
When defining a struct, and its' methods, the following order should be used in the file:
 1) Struct definition
 2) Struct factory
 3) Struct exported methods (alphabetically ordered)
 4) Struct non-exported methods (alphabetically ordered)

Example:
```go
type Foo struct {
    a int
}

// NewFoo creates a Foo object
func NewFoo(a int) Foo {
    return Foo{a: a}
}

func (f Foo) String() string {
    return fmt.Sprintf("Foo#%d", f.a)
}

// Add a different Foo object to this one and returns the resultant Foo.
func (f Foo) Add(g Foo) Foo {
    return Foo{a: f.a + g.a}
}

// zero will zero out the Foo object and return the resultant Foo.
func (f Foo) zero() Foo {
    return Foo{a: 0}
}
``` 
 
## Synchronous functions
Prefer synchronous functions - functions which return their results directly or finish any callbacks or channel ops
before returning - over asynchronous ones.

Synchronous functions keep goroutines localized within a call, making it easier to reason about their lifetimes and
avoid leaks and data races. They're also easier to test: the caller can pass an input and check the output without the
need for polling or synchronization.

If callers need more concurrency, they can add it easily by calling the function from a separate goroutine. But it is
quite difficult - sometimes impossible - to remove unnecessary concurrency at the caller side.

## Third-party packages
The inclusion of additional external packages will be considered with extreme caution and will require solid
justification. "I'm not able to write effective tests without package XYZ" is not a good reason.

## Variable names
Variable names, including struct properties, should short, but not at the cost of being non-descriptive.  We are not
paying per keypress!

Do not use variable names which collide with imported package. Example:
```go
// Bad
import "database/sql"

// ...

sql := getAssetSQL()
```

```go
// Good
import "database/sql"

// ...

assetSQL := getAssetSQL()
```
