#!/bin/bash

# Retrieve files based on the extension from the current directory.
# Args:
#   $0: Source directory to search in.
#   $1: The extension type, wild cards can be included.

src_dir=$0
ext_type=$1

find $src_dir -type f -name "$ext_type"