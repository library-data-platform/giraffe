Giraffe
=======

Copyright (C) 2019 The Open Library Foundation.  This software is
distributed under the terms of the Apache License, Version 2.0.  See the
file
[LICENSE](https://github.com/folio-labs/giraffe/blob/master/LICENSE) for
more information.


Overview
--------

Giraffe is a tool for creating visualizations of Okapi logs such as call
graphs.

![Giraffe example](https://github.com/folio-labs/giraffe/blob/master/okapi-example.png "Giraffe example")


System requirements
-------------------

* Linux, macOS, possibly Windows
* Go 1.10 or later
* Graphviz

Giraffe has not been tested in Windows, but it may work normally or
using the "Two Step" method below.


Installing Giraffe
------------------

First ensure that the `GOPATH` environment variable specifies a path
that can serve as your Go workspace directory, the place where this
software and other Go packages will be installed.  For example, to set
it to `$HOME/go`:

```shell
$ export GOPATH=$HOME/go
```

Then to download and compile the software (or to retrieve the latest
updates):

```shell
$ go get -u github.com/folio-labs/giraffe/...
```

The compiled executable file `giraffe` should appear in `$GOPATH/bin/`.


Running Giraffe
---------------

To generate a call graph in PDF format from part of an Okapi log stored
in `okapi-part.log`:

```shell
$ giraffe call -i okapi-part.log -o okapi-part.pdf
```

Other output formats, PNG, JPEG and DOT, are supported via the `-T`
flag, for example:

```shell
$ giraffe call -i okapi-part.log -o okapi-part.png -T png
```

For more information:

```shell
$ giraffe help
```


Running Giraffe in Two Steps
----------------------------

If the instructions above do not work, for instance in Windows, try
running Giraffe in two steps using dot:

```shell
giraffe call -i test.log -o test.dot -T dot
```
```shell
dot -o test.pdf -T pdf test.dot
```


