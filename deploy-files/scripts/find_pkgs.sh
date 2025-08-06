#!/bin/bash

# script used to retrieve all pkg files in a given directory.
# args:
#   - 

src_dir=$0

find $src_dir -type f -name "*.pkg"