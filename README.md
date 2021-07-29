Giraffe
=======

Copyright (C) 2019-2021 The Open Library Foundation.  This software is
distributed under the terms of the Apache License, Version 2.0.  See
the file
[LICENSE](https://github.com/folio-labs/giraffe/blob/master/LICENSE)
for more information.


Overview
--------

Giraffe is a tool for creating visualizations of Okapi logs such as call
graphs.

![Giraffe example](https://github.com/folio-labs/giraffe/blob/master/example.png "Giraffe example")


System requirements
-------------------

* Linux, macOS, possibly Windows
* Go 1.16 or later
* Graphviz

Giraffe has not been tested in Windows, but it may work normally or
using the "Two Step" method below.


Installing Giraffe
------------------

First set the `GOPATH` environment variable to specify a path that can
serve as the build workspace for Go, e.g.:

```bash
$ export GOPATH=$HOME/go
```

Then:

```bash
go build ./cmd/giraffe
```

This builds the `giraffe` executable in the current directory.


Running Giraffe
---------------

To generate a call graph in PDF format from part of an Okapi log stored
in `sample.log`:

```shell
$ giraffe call -i sample.log -o sample.pdf
```

Other output formats, PNG, JPEG and DOT, are supported via the `-T`
flag, for example:

```shell
$ giraffe call -i sample.log -o sample.png -T png
```

Giraffe can highlight response times that meet a specified threshold,
given in milliseconds.  For example, to highlight response times of at
least 250 ms:

```shell
$ giraffe call -i sample.log -o sample.pdf -rstime 250
```

For more information about command line flags:

```shell
$ giraffe help
```


Running Giraffe in Two Steps
----------------------------

If the instructions above do not work, for instance in Windows, try
running Giraffe in two steps using dot:

```shell
giraffe call -i sample.log -o sample.dot -T dot
```
```shell
dot -o sample.pdf -T pdf sample.dot
```


