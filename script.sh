#!/bin/bash

# This scripts takes a root of a directory and a file extension as input
# It then finds all the files with that extension and merge into a single file

path=$1
ext=$2
out=$3

# Check if the path is a directory
if [ ! -d $path ]; then
    echo "The path is not a directory"
    exit 1
fi

# Traverse the directory and find all the files with the given extension
# Store the paths of the files in a array
files=()
while IFS=  read -r -d $'\0'; do
    files+=("$REPLY")
done < <(find $path -type f -name "*.$ext" -print0)

# Save the content of all the files in a single file, each file separated by a newline and a comment with path
for file in "${files[@]}"; do
    echo "# $file" >> $out
    cat $file >> $out
    echo "" >> $out
done
