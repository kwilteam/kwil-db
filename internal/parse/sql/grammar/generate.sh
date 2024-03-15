#!/bin/sh

set -e

target=${1:-Go}
package=${2:-sqlgrammar}
output_dir=${3:-../gen}

antlr_bin=antlr-4.13.1-complete.jar

if [ ! -f $antlr_bin ]; then
    echo "Downloading antlr4 jar file..."
    curl -O https://www.antlr.org/download/${antlr_bin}
fi

alias antlr4='java -Xmx500M -cp "./${antlr_bin}:$CLASSPATH" org.antlr.v4.Tool'
antlr4 -Dlanguage="${target}" -visitor -no-listener -package "${package}" -o "${output_dir}" *.g4