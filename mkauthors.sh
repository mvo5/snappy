#!/bin/bash
set -e

# debugging if anything fails is tricky as dh-golang eats up all output
# uncomment the lines below to get a useful trace if you have to touch
# this again (my advice is: DON'T)
#set -x
#logfile=/tmp/mkauthors.log
#exec >> $logfile 2>&1
#echo "env: $(set)"
#echo "mkauthors.sh run from: $0"
#echo "pwd: $(pwd)"

# we have two directories we need to care about:
# - our toplevel pkg builddir which is where "mkauthors.sh" is located
#   and where "snap-confine" expects its cmd/VERSION file
# - the GO_GENERATE_BUILDDIR which may be the toplevel pkg dir. but
#   during "dpkg-buildpackage" it will become a different _build/ dir
#   that dh-golang creates and that only contains a subset of the
#   files of the toplevel buildir.
PKG_BUILDDIR=$(dirname "$0")
GO_GENERATE_BUILDDIR="$(pwd)"

# run from "go generate" adjust path
if [ "$GOPACKAGE" = "cmd" ]; then
    GO_GENERATE_BUILDDIR="$(pwd)/.."
fi

# Let's try to derive authors from git
if ! command -v git >/dev/null; then
    exit
fi

# see if we are in a git branch, if not, do nothing
if [ ! -d .git ]; then
    exit
fi

raw_authors="$(git shortlog -s|sort -n|tail -n14|cut -f2)"
authors=""
while read -r author; do
    authors="$authors \"$author\","
done <<< "$raw_authors"


cat <<EOF > "$GO_GENERATE_BUILDDIR/cmd/snap/cmd_blame_generated.go"
package main

// generated by mkauthors.sh; do not edit

func init() {
	authors = []string{"Mark Shuttleworth", "Gustavo Niemeyer", $authors}
}
EOF

go fmt $GO_GENERATE_BUILDDIR/cmd/snap/cmd_blame_generated.go >/dev/null
