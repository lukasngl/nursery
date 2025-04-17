#!/usr/bin/env sh

while [ $# -gt 0 ]; do
	just --unstable --fmt --justfile "$1"
	shift
done
