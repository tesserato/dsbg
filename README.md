# dead-simple-blog-generator

A dead simple blog generator

# Description

This programs concatenates a codebase in a single markdown file, to be used in LLMs

Given a directory, it generates a markdown file starting with a tree of the codebase and followed by the contents of all files concatenated in code blocks.

accepts a directory as an argument, a name for the output file and a list of path patterns to ignore.

# Principles

out of the box

DRY(ish)

Explicit


# Comand line arguments

Directory to scan (default: current directory)

Output file name (default: codebase.md)

List of Path patterns to ignore (default: none)