pkg-updated for FreeBSD
=====================

pkg-updated is a daemon/wrapper for automated binary package update.
The aim of this project is to provide a simple and reliable pkg update daemon for FreeBSD.

The daemon is written in go and programmed with the literate programming method.
Code and documentation are sourced in one file and need to extract/weaved before use.

PKG (aka. pkgng) is the default binary package management software for FreeBSD.

Features:

- Configurable scheduler for recurring times
- Archive packages which need to upgrade (for rollback)
- Restart services if an enabled running service was updated
- Rollback updated packages on failed service restart (not finished)


How to Start
---------------------
1\. Install git client

2\. Download the code:

```bash
git clone https://github.com/scd-systems/pkg-updated.git
```

3\. Create the documentation:

```bash
cd src;
make pdf
```

4\. Open and read the pkg-updated.pdf

5\. Modify/Extend/Change the source.
Use your favorite tex/latex editor.
Open the `pkg-updated.nw` file
If you are done, re-create the documentation, the code or both:

```bash
cd src;
make
```

6\. Fetch extra sqlite go package
```bash
setenv GOPATH .
go install github.com/mattn/go-sqlite3 
```

7\. Run
```bash
go run pkg-updated.go
```


Installation
---------------------
Please read documentations/install.howto

Copyright
---------------------
For copyright information to this Project (pkg-updated), please see the file LICENSE in this directory.
