#!/bin/bash

src_dir=$1
pattern=$2

find $src_dir -type f -name "*$pattern*"