#!/bin/sh

antlr_bin=antlr-4.12.0-complete.jar

alias antlr4='java -Xmx500M -cp "./${antlr_bin}:$CLASSPATH" org.antlr.v4.Tool'
antlr4 -Dlanguage=Go -no-visitor -package sqlite *.g4